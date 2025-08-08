package internals

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	port        = 4000
	broadcastIP = "192.168.1.255:4000" // FIXME: hardcoded, need to get the subnet from something like ifconfig
)

func Serve() {
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	broadcastAddr, err := net.ResolveUDPAddr("udp4", broadcastIP)
	if err != nil {
		panic(err)
	}

	found := make(chan struct{})

	go func() {
		for {
			select {
			case <-found:
				return
			default:
				_, err := conn.WriteTo([]byte("lethergo"), broadcastAddr)
				if err != nil {
					fmt.Println("Broadcast error:", err)
				}
				time.Sleep(3 * time.Second)
			}
		}
	}()

	buf := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		msg := strings.TrimSpace(string(buf[:n]))

		if msg == "lethergo" {
			fmt.Printf("found: %s!\n", addr)
			close(found)

		} else {
			fmt.Printf("bs from: %s: %s\n", addr, msg)
		}
	}
}
