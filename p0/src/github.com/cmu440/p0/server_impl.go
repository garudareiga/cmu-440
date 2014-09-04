// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"fmt"
	"net"
	"os"
	//"strings"
)

type echoClient struct {
	conn net.Conn
	ch   chan string
}

type multiEchoServer struct {
	// TODO: implement this!
	host    string
	clients map[int]*echoClient // Id -> Client
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	ptrServer := &multiEchoServer{
		host:    "localhost",
		clients: make(map[int]*echoClient),
	}
	return MultiEchoServer(ptrServer)
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)) // create a server
	if err != nil {
		fmt.Println("Error on listen: ", err)
		os.Exit(-1)
	}
	fmt.Printf("Server is running at %s:%d\n", mes.host, port)

	go handleServer(ln, mes.clients)

	return nil
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	for _, c := range mes.clients {
		c.conn.Close()
	}
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	return len(mes.clients)
}

// TODO: add additional methods/functions below!
func handleServer(ln net.Listener, clients map[int]*echoClient) {
	//eclChan := make(chan map[int]*echoClient, 1)
	//eclChan <- clients

	msgChan := make(chan string)
	go handleMessage(msgChan, clients)

	i := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error on accept: ", err)
			continue
		}

		//go handleConnection(i, conn, eclChan)
		go handleConnection(i, conn, clients, msgChan)
		i++
	}
}

//func handleConnection(i int, conn net.Conn, clients map[int]*echoClient) {
//func handleConnection(i int, conn net.Conn, eclChan chan map[int]*echoClient) {
func handleConnection(i int, conn net.Conn, clients map[int]*echoClient, msgChan chan<- string) {
	fmt.Printf("Client %d: %v <-> %v\n", i, conn.LocalAddr(), conn.RemoteAddr())

	ptrEchoClient := &echoClient{
		conn: conn,
		ch:   make(chan string, 100),
	}

	//clients := <-eclChan
	clients[i] = ptrEchoClient
	//eclChan <- clients

	go echo(ptrEchoClient)
	defer ptrEchoClient.conn.Close()

	rb := bufio.NewReader(conn)
	for {
		msg, e := rb.ReadString('\n')
		if e != nil {
			break
		}
		msgChan <- msg
		//clients = <-eclChan
		//eclChan <- clients
	}

	//clients = <-eclChan
	delete(clients, i)
	//eclChan <- clients
	fmt.Printf("%d: closed\n", i)
}

func echo(client *echoClient) {
	for {
		msg := <-client.ch
		_, err := client.conn.Write([]byte(msg))
		if err != nil {
			break
		}
	}
}

func handleMessage(msgChan <-chan string, clients map[int]*echoClient) {
	for {
		msg := <-msgChan
		for _, echoClient := range clients {
			select {
			case echoClient.ch <- msg:
			default:
			}
		}
	}
}
