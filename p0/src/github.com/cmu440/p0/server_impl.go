// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"fmt"
	"net"
)

type echoClient struct {
	conn net.Conn
	ch   chan string
}

type multiEchoServer struct {
	// TODO: implement this!
	host    string
	eclChan chan map[int]*echoClient
	ln      net.Listener
	stop    chan bool
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	ptrServer := &multiEchoServer{
		host:    "localhost",
		eclChan: make(chan map[int]*echoClient, 1),
		stop:    make(chan bool),
	}
	return MultiEchoServer(ptrServer)
}

func (mes *multiEchoServer) Start(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)) // create a server
	mes.ln = ln
	if err != nil {
		fmt.Println("Error on listen: ", err)
		return err
	}
	fmt.Printf("Server is running at %s:%d\n", mes.host, port)

	go mes.serve()

	return nil
}

func (mes *multiEchoServer) Close() {
	for _, client := range <-mes.eclChan {
		client.conn.Close()
	}
	mes.ln.Close()
	close(mes.stop)
}

func (mes *multiEchoServer) Count() int {
	clients := <-mes.eclChan
	count := len(clients)
	mes.eclChan <- clients
	return count
}

// TODO: add additional methods/functions below!
func (mes *multiEchoServer) serve() {
	msgChan := make(chan string)
	go handleMessage(msgChan, mes.eclChan)

	clients := make(map[int]*echoClient)
	mes.eclChan <- clients

	i := 0
	for {
		conn, err := mes.ln.Accept()
		if err != nil {
			select {
			case <-mes.stop:
				return
			default:
			}
			fmt.Println("Error on accept: ", err)
			continue
		}

		go handleConnection(i, conn, mes.eclChan, msgChan)
		i++
	}
}

func handleConnection(i int, conn net.Conn, eclChan chan map[int]*echoClient, msgChan chan<- string) {
	fmt.Printf("Client %d: %v <-> %v\n", i, conn.LocalAddr(), conn.RemoteAddr())

	ptrEchoClient := &echoClient{
		conn: conn,
		ch:   make(chan string, 100),
	}

	clients := <-eclChan
	clients[i] = ptrEchoClient
	eclChan <- clients

	go echo(ptrEchoClient)
	defer ptrEchoClient.conn.Close()

	rb := bufio.NewReader(conn)
	for {
		msg, e := rb.ReadString('\n')
		if e != nil {
			break
		}
		msgChan <- msg
	}

	clients = <-eclChan
	delete(clients, i)
	eclChan <- clients
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

func handleMessage(msgChan <-chan string, eclChan chan map[int]*echoClient) {
	for {
		msg := <-msgChan
		clients := <-eclChan
		for _, echoClient := range clients {
			select {
			case echoClient.ch <- msg:
			default: // discard value if channel is full
			}
		}
		eclChan <- clients
	}
}
