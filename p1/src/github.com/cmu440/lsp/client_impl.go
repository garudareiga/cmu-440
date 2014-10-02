// Contains the implementation of a LSP client.

package lsp

import (
	"encoding/json"
	"errors"
	//	"fmt"
	"github.com/cmu440/lspnet"
	"log"
	//	"time"
)

type client struct {
	// TODO: implement this!
	hostport string
	params   *Params
	connID   int
	seqNum   int
	conn     *lspnet.UDPConn

	connected bool
	connOpen  chan struct{}
	connClose chan struct{}
	connMsg   *Message

	dataMsgTo chan *Message
	dataMsgFr chan *Message

	dataMsgTokens chan struct{}

	seqNumToActChan    chan map[int]*Activity
	seqNumToAckMsgChan chan map[int]*Message

	dataWrWindow chan *SlideWindow
	dataRdWindow chan *SlideWindow
}

// NewClient creates, initiates, and returns a new client. This function
// should return after a connection with the server has been established
// (i.e., the client has received an Ack message from the server in response
// to its connection request), and should return a non-nil error if a
// connection could not be made (i.e., if after K epochs, the client still
// hasn't received an Ack message from the server in response to its K
// connection requests).
//
// hostport is a colon-separated string identifying the server's host address
// and port number (i.e., "localhost:9999").
func NewClient(hostport string, params *Params) (Client, error) {
	c := &client{
		hostport: hostport,
		params:   params,
		connID:   -1,
		seqNum:   0,

		conn:      nil,
		connected: false,
		connOpen:  make(chan struct{}),
		connClose: make(chan struct{}),
		connMsg:   NewConnect(),

		dataMsgTo: make(chan *Message),
		dataMsgFr: make(chan *Message),

		dataMsgTokens: make(chan struct{}, params.WindowSize),

		seqNumToActChan:    make(chan map[int]*Activity),
		seqNumToAckMsgChan: make(chan map[int]*Message),

		dataWrWindow: make(chan *SlideWindow, 1),
		dataRdWindow: make(chan *SlideWindow, 1),
	}

	raddr, err := lspnet.ResolveUDPAddr("udp", hostport)
	if err != nil {
		return nil, err
	}
	conn, err := lspnet.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	c.dataWrWindow <- &SlideWindow{
		seqNumLower: 1,
		seqNumUpper: 1 + c.params.WindowSize,
	}
	c.dataRdWindow <- &SlideWindow{
		seqNumLower: 1,
		seqNumUpper: 1 + c.params.WindowSize,
	}

	go c.onWrite()
	go c.onRead()
	go c.epoch()
	go c.connect()

	for {
		select {
		case <-c.connOpen:
			c.connected = true
			return c, nil
		case <-c.connClose:
			c.connected = false
			return nil, errors.New("connection not yet established")
		}
	}

	return nil, errors.New("not yet implemented")
}

func (c *client) ConnID() int {
	//return -1
	return c.connID
}

func (c *client) Read() ([]byte, error) {
	// TODO: remove this line when you are ready to begin implementing this method.
	//select {} // Blocks indefinitely.
	for {
		select {
		case <-c.connClose:
			return nil, errors.New("connection closed")
		case message := <-c.dataMsgFr:
			return message.Payload, nil
		}
	}
	return nil, errors.New("not yet implemented")
}

func (c *client) Write(payload []byte) error {
	// Need a token to write a new message
	<-c.dataMsgTokens

	// Write a new message
	newMsg := NewData(c.connID, c.seqNum, payload)
	c.seqNum++
	signal := make(chan struct{})
	go c.writeData(newMsg, signal)

	for {
		select {
		case <-c.connClose:
			return errors.New("connection closed")
		case <-signal:
			return nil
		}
	}
	return errors.New("not yet implemented")
}

func (c *client) Close() error {
	close(c.connClose)
	return nil
	//return errors.New("not yet implemented")
}

func (c *client) connect() {
	writeBytes, _ := json.Marshal(c.connMsg)
	_, err := c.conn.Write(writeBytes)
	if err != nil {
		log.Println("error: send connect message")
		c.Close()
	}
}

func (c *client) writeData(message *Message, signal chan struct{}) {
	act := &Activity{
		message: message,
		signal:  signal,
		done:    false,
	}
	activities := <-c.seqNumToActChan
	activities[message.SeqNum] = act
	c.seqNumToActChan <- activities

	// add data message to send channel
	c.dataMsgTo <- message
}

// Send data message
func (c *client) onWrite() {
	for {
		select {
		case <-c.connClose:
			return
		case message := <-c.dataMsgTo:
			writeBytes, _ := json.Marshal(message)
			_, err := c.conn.Write(writeBytes)
			if err != nil {
				log.Printf("error: send message %s\n", writeBytes)
				continue
			}
		}
	}
}

func (c *client) onRead() {
	for {
		select {
		case <-c.connClose:
			return
		}

		var readBytes []byte
		n, err := c.conn.Read(readBytes)
		if err == nil {
			continue
		}

		var message *Message
		json.Unmarshal(readBytes[:n], message)

		if message.Type == MsgAck { // Ack
			if c.connected == false {
				go c.onConnAckMsg(message)
			} else {
				go c.onDataAckMsg(message)
			}
		} else { // Data
			go c.onData(message)
		}
	}
}

func (c *client) onConnAckMsg(message *Message) {
	c.connID = message.ConnID
	c.connected = true
	c.connOpen <- struct{}{}
}

func (c *client) onDataAckMsg(message *Message) {
	activities := <-c.seqNumToActChan
	if act, ok := activities[message.SeqNum]; ok {
		act.signal <- struct{}{}
		act.done = true
	}
	c.seqNumToActChan <- activities
	go c.checkDataWindow(message.SeqNum)
}

func (c *client) onData(message *Message) {
	seqNum := message.SeqNum

	sw := <-c.dataRdWindow
	seqNumLower := sw.seqNumLower
	seqNumUpper := sw.seqNumUpper
	c.dataRdWindow <- sw

	// Check if duplicate
	if seqNum < seqNumLower || seqNum >= seqNumUpper {
		return
	}

	seqNumToAckMsg := <-c.seqNumToAckMsgChan
	_, ok := seqNumToAckMsg[seqNum]
	if ok {
		c.seqNumToAckMsgChan <- seqNumToAckMsg
		return
	}

	// Create Ack message
	ackMsg := NewAck(c.ConnID(), seqNum)
	seqNumToAckMsg[message.SeqNum] = ackMsg
	c.seqNumToAckMsgChan <- seqNumToAckMsg

	go c.checkAckWindow(seqNum)
	c.dataMsgFr <- message
}

func (c *client) epoch() {

}

func (c *client) checkDataWindow(seqNum int) {
	// decrease seqNumLower and free tokens
}

func (c *client) checkAckWindow(seqNum int) {
	//
}
