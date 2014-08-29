// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"fmt"
	"net"
	"os"
)

type multiEchoServer struct {
	// TODO: implement this!
	host string
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	//return nil
	ptrServer := &multiEchoServer{ host:"localhost" }
	return MultiEchoServer(ptrServer)
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error on listen: ", err)
		os.Exit(-1)
	}

	for {
		fmt.Println("Waiting for a connection via Accept")
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error on accept: ", err)
			os.Exit(-1)
		}

		myconn := &myConn { conn:conn, prefix:"Server" }
		go handleRequest(myconn)
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
type myConn struct {
	conn net.Conn
	prefix string
}

func handleRequest(myconn *myConn) {
	fmt.Println("Reading once from connection")

	var buf [1024]byte
	n, err := myconn.conn.Read(buf[:])
	if err != nil {
		fmt.Println("Error on read: ", err)
		os.Exit(-1)
	}
	
	fmt.Println(myconn.prefix, ":", string(buf[0:n]))
	//myconn.conn.Close()
}