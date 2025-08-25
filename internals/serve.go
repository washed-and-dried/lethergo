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
	addr  net.Addr
	state int
	id    int64
}

const (
	BROADCASTING = iota
	FOUND
)

const (
	port     = ":4000"
	TCP_PORT = ":4001"
)

// var broadcastIP = GetLocalIP() + port
var broadcastIP = "255.255.255.255" + port
// var broadcastIP = "172.17.0.255" + port

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
		for {
			select {
			case chanVal := <- state:
				{
					switch chanVal.state {
					case FOUND:
						{
							_, err := conn.WriteTo(fmt.Appendf(nil, "shegooner %d", chanVal.id), broadcastAddr)
							if err != nil {
								fmt.Println("Broadcast error:", err)
								return
							}

							wg.Add(1)
							go listenTCP(chanVal.addr, &wg)
						}

					case BROADCASTING:
						{
							_, err := conn.WriteTo(fmt.Appendf(nil, "lethergo %d", id), broadcastAddr)
							fmt.Printf("broadcasting id: %d\n", id)
							if err != nil {
								fmt.Println("Broadcast error:", err)
								return
							}
							time.Sleep(3 * time.Second)
						}
					}
				}
			}
		}
	}()

	state <- GoMsg{
		state: BROADCASTING,
	}

	buf := make([]byte, 1024)
	found := false

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
					id:    id,
					addr:  addr,
				}
				found = true
			} else if msg[1] == fmt.Sprintf("%d", id) && msg[0] == "shegooner" {
				fmt.Printf("shegooner found: %s\n", addr.String())
				wg.Add(1)
				go sendTCP(addr, &wg)
				return
			} else if !found {
				state <- GoMsg{
					state: BROADCASTING,
				}
			}
		}
	}()
	wg.Wait()
	fmt.Println("Bye")
}

func listenTCP(addr net.Addr, wg *sync.WaitGroup) {
	defer wg.Done()

	listenAddr, err := net.ResolveTCPAddr("tcp", GetLocalIP()+TCP_PORT)

	connListener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	defer connListener.Close()

	fmt.Printf("Listening to connections on port %s", TCP_PORT)

	for {
		conn, err := connListener.Accept()
		if err != nil {
			fmt.Printf("Could not accept the connection to the server: %s\n", err)
			return
		}

		fmt.Printf("Successfully connected to client: %s\n", conn.RemoteAddr().String())

		wg.Add(1)
		go handleConn(conn, wg)
	}
}

func sendTCP(addr net.Addr, wg *sync.WaitGroup) {
	defer wg.Done()

	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		fmt.Println("Could not cast %T to TCPAddr\n", addr)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Printf("Dial failed: %s\n", err)
	}

	handleConn(conn, wg)
}

func handleConn(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	// FIXME: handleConn
	panic("handleConn")
}
