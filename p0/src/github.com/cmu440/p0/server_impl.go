// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"fmt"
	"net"
	"os"
)

type multiEchoServer struct {
	// TODO: implement this!
	host     string
	clientId int
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	//return nil
	ptrServer := &multiEchoServer{host: "localhost"}
	return MultiEchoServer(ptrServer)
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error on listen: ", err)
		os.Exit(-1)
	}

	msgchan := make(chan string) // Message channel
	addchan := make(chan Client) // Add-Client channel
	delchan := make(chan Client) // Del-Client channel

	go onMessage()

	for {
		fmt.Println("Waiting for a connection via Accept")
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error on accept: ", err)
			continue
		}

		go onConnection(conn, msgchan)
	}
	return nil
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	return -1
}

// TODO: add additional methods/functions below!
type Client struct {
	conn net.Conn
	ch   <-chan string
}

func onMessage(msgchan <-chan string, addchan <-chan Client, delchan <-chan Client) {
	clmap := make(map[net.Conn]chan<- string)

	for {
		select {
		case msg := <-msgchan:
			fmt.Printf("New echo message: " + msg)
			for _, ch := range clmap {
				// go func(mch chan<- string) { mch <- msg }();
				ch <- msg
			}
		case client := <-addchan:
			fmt.Printf("New connection: " + client.conn)
			clmap[client.conn] = client.ch
		case client := <-delchan:
			fmt.Printf("Lost connection: " + client.conn)
			delete(clmap, client.conn)
		}
	}
}

func onConnection(s net.Conn, msgchan chan<- string, addchan chan<- Client, delchan chan<- Client) {
	defer s.Close()

	fmt.Printf("%d: %v <-> %v\n", i, s.LocalAddr(), s.RemoteAddr())

	client := Client{conn: s, ch: make(chan string)}
	addchan <- client

	go onEcho(client.conn, client.ch)

	b := bufio.NewReader(s)
	for {
		//line, e := b.ReadBytes('\n')
		line, e := b.ReadString('\n')
		if e != nil {
			break
		}
		msgchan <- line
		//s.Write(line)
	}

	delchan <- client
	fmt.Printf("%d: closed\n", i)
}

func onEcho(s net.Conn, ch <-chan string) {
	b := bufio.NewWriter(s)
	for {
		msg <- ch
		_, err := b.WriteString(msg)
		if err != nil {
			break
		}
	}
}
