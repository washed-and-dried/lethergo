package internals

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
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

type GoMsg struct {
	state int
}

const (
	BROADCASTING = iota
	FOUND
)

const (
	port = ":4000"
)

// var broadcastIP = GetLocalIP() + port
var broadcastIP = "255.255.255.255" + port

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

	state := make(chan GoMsg)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		var chanVal GoMsg
		for chanVal = range state {
			switch chanVal.state {
			case FOUND:
				return // FIXME: breaking for now
			case BROADCASTING:
				{
					_, err := conn.WriteTo(fmt.Appendf(nil, "lethergo %d", id), broadcastAddr)
					fmt.Println("broadcasting id: %d", id)
					if err != nil {
						fmt.Println("Broadcast error:", err)
					}
					time.Sleep(3 * time.Second)
				}
			}
		}
	}()

	state <- GoMsg{
		state: BROADCASTING,
	}

	buf := make([]byte, 1024)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				fmt.Println("Read error:", err)
				continue
			}

			msg := strings.Split(strings.TrimSpace(string(buf[:n])), " ")

			if msg[0] == "lethergo" && msg[1] != fmt.Sprintf("%d", id) {
				fmt.Printf("found: %s!\n", addr)
				fmt.Printf("%s\n", msg)
				// close(state)
				state <- GoMsg{
					state: FOUND,
				}
				break
			} else {
				state <- GoMsg{
					state: BROADCASTING,
				}
			}
		}
	}()
	wg.Wait()
	fmt.Println("Bye")
}

func handleConn(addr net.Addr) {
}
