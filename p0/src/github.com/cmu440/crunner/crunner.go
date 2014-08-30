package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"bufio"
	"strings"
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

	ch := make(chan string)

	// handle outbound message
	go func(conn net.Conn, ch <-chan string) {
		b := bufio.NewWriter(conn)
		for {
			select {
			case msg := <- ch:
				_, err := b.WriteString(msg + "\n")
				b.Flush()
				if err != nil {
					fmt.Println("Error on write: ", err)
					os.Exit(-1)
				}	
				fmt.Println("Write " + msg + " to connection")
			}
		}
	}(conn, ch)

	// handle inbound message
	go func(conn net.Conn) {
		b := bufio.NewReader(conn)
		for {
			line, err := b.ReadString('\n')
			if err != nil {
				break
			}
			fmt.Println("Echo back: ", strings.TrimRight(string(line), "\n"))
		}
	}(conn)

	for {
		var input string
		fmt.Scanln(&input)	

		if input == "done" {
			break
		}	

		ch <- input
	}

	conn.Close()
	fmt.Println("done")
	//fmt.Println("Not implemented.")
}
