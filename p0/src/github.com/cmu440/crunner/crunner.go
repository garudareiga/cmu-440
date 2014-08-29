package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	defaultHost = "localhost"
	defaultPort = 9999
)

// To test your server implementation, you might find it helpful to implement a
// simple 'client runner' program. The program could be very simple, as long as
// it is able to connect with and send messages to your server and is able to
// read and print out the server's echoed response to standard output. Whether or
// not you add any code to this file will not affect your grade.
func main() {
	conn, err := net.Dial("tcp", defaultHost + ":" + strconv.Itoa(defaultPort))
	if err != nil {
		fmt.Println("Error on dial: ", err)
		os.Exit(-1)
	}

	fmt.Println("Write once to connection")
	buf := []byte("Hello!") 
	n, err := conn.Write(buf)
	if err != nil {
		fmt.Println("Error on write: ", err)
		os.Exit(-1)
	}

	fmt.Println("Write ", buf[:n])

	conn.Close()
	//fmt.Println("Not implemented.")
}
