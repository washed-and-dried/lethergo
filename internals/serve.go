package internals

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	// remove the last octect and insert the 255 for broadcasting
	fullIP := localAddress.IP.String()
	idx := strings.LastIndex(fullIP, ".")
	broadcast := fullIP[:idx+1] + "255"

	return broadcast
}

const (
	port = ":4000"
)

var broadcastIP = GetLocalIP() + port

func Serve() {
	fmt.Println(broadcastIP)
	conn, err := net.ListenPacket("udp4", port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	broadcastAddr, err := net.ResolveUDPAddr("udp4", broadcastIP)
	if err != nil {
		panic(err)
	}

	var id int64 = rand.Int63()

	found := make(chan struct{})

	go func() {
		for {
			select {
			case <-found:
				return
			default:
				_, err := conn.WriteTo(fmt.Appendf(nil, "lethergo %d", id), broadcastAddr)
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

		msg := strings.Split(strings.TrimSpace(string(buf[:n])), " ")

		if msg[0] == "lethergo" && msg[1] != fmt.Sprintf("%d", id) {
			fmt.Printf("found: %s!\n", addr)
			close(found)
			break
		} else {
			fmt.Printf("bs from: %s: %s\n", addr, msg)
		}
	}

	fmt.Println("Bye")
}
