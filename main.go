package main

import (
	"fmt"
	"net"
	"strings"
)

// Similiarly to the K/V database I have used method of passing the function in the
// first byte of the message transmission
// S will subscribe to any broadcast messages
// B:MESSAGE will broadcast the message to any connected clients

type tcpclient struct {
	cn net.Conn
}

type udpclient struct {
	cn   *net.UDPConn
	addr *net.UDPAddr
}

// Create channels required
type hub struct {
	tcpclients map[*tcpclient]bool
	udpclients map[*udpclient]bool
	broadcast  chan []byte
	tregister  chan *tcpclient
	uregister  chan *udpclient
}

// Create hub to manage the clients and messages
var h = hub{
	broadcast:  make(chan []byte),
	tregister:  make(chan *tcpclient),
	tcpclients: make(map[*tcpclient]bool),
	udpclients: make(map[*udpclient]bool),
}

func main() {
	// Creates a listener on TCP
	l, err := net.Listen("tcp", "127.0.0.1:1111")
	if err != nil {
		fmt.Print(err)
	}
	// Create a listener on UDP
	addr := net.UDPAddr{
		Port: 1111,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	// Run the TCP and UDP as go functions
	go acceptTCP(l)
	go acceptUDP(ser)
	h.run()
}

// TCP Processor
func acceptTCP(l net.Listener) {
	defer l.Close()
	for {
		// Accept the connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Create a 2k buffer to hold the message recieved
		m := make([]byte, 2048)
		n, _ := conn.Read(m)
		// Handle if subscribing to messages or broadcasting a message
		f := m[:1]
		switch strings.ToUpper(string(f)) {
		case "B":
			// Pass through the message to be broadcast
			h.broadcastMessage(m[2:n])
		case "S":
			// Pass through the tcp connection to add to the connection map
			h.tcpsubChan(conn)
		}
	}
}

// UDP Processor
func acceptUDP(l *net.UDPConn) {
	//  *cacheCreate a 2k buffer to hold the message recieved
	m := make([]byte, 2048)
	for {
		// Handle UDP slightly differently from TCP
		n, addr, err := l.ReadFromUDP(m)
		if err != nil {
			fmt.Printf("Some error %v\n", err)
		}
		// Handle if subscribing to messages or broadcasting a message
		f := m[:1]
		switch strings.ToUpper(string(f)) {
		case "B":
			// Pass through the message to be broadcast
			h.broadcastMessage(m[2:n])
		case "S":
			// Pass through the udp connection and return address as udp connections
			// create a seperate return port and add to the connection map
			h.udpsubChan(l, addr)
		}
	}
}

// Subscribe the UDP connection
func (h *hub) udpsubChan(l *net.UDPConn, a *net.UDPAddr) {
	c := &udpclient{
		cn:   l,
		addr: a,
	}
	h.uregister <- c
}

// Subscribe the TCP connection
func (h *hub) tcpsubChan(a net.Conn) {
	c := &tcpclient{
		cn: a,
	}
	h.tregister <- c
}

// Pass the message to the broadcast channel
func (h *hub) broadcastMessage(m []byte) {
	h.broadcast <- m
}

// Go routine that constantly monitors the channels and
// processes when recieving through each channel
func (h *hub) run() {
	for {
		select {
		// Add to TCP client Map
		case c := <-h.tregister:
			h.tcpclients[c] = true
			break
		// Add to UDP client map
		case c := <-h.uregister:
			h.udpclients[c] = true
			break
		// Broadcast message to all clients
		case m := <-h.broadcast:
			h.bMessage(m)
			break
		}
	}
}

// Loop through all connections and broadcast the message
func (h *hub) bMessage(m []byte) {
	for c := range h.tcpclients {
		_, err := c.cn.Write(m)
		if err != nil {
			delete(h.tcpclients, c)
		}
	}
	for c := range h.udpclients {
		_, err := c.cn.WriteToUDP(m, c.addr)
		if err != nil {
			delete(h.udpclients, c)
		}
	}
}
