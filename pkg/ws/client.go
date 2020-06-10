package ws

import (
	"chat/pkg/model"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"time"
)


const channelBufSize = 100

// Chat client.
type Client struct {
	identity     *model.User
	ws     *websocket.Conn
	server *Server
	ch     chan *Message
	doneCh chan bool
}

// Create new chat client.
func NewClient(ws *websocket.Conn, server *Server, identity *model.User) *Client {
	if ws == nil {
		panic("ws cannot be nil")
	}

	if server == nil {
		panic("server cannot be nil")
	}

	ch := make(chan *Message, channelBufSize)
	doneCh := make(chan bool)
	return &Client{identity, ws, server, ch, doneCh}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Write(msg *Message) {
	fmt.Println(msg)
	select {
	case c.ch <- msg:
	default:
		c.server.Del(c)
		err := fmt.Errorf("client %d is disconnected.", c.identity.Id)
		c.server.Err(err)
	}
}

func (c *Client) Done() {
	c.doneCh <- true
}

// Listen Write and Read request via chanel
func (c *Client) Listen() {
	go c.listenWrite()
	c.listenRead()
}

// Listen write request via chanel
func (c *Client) listenWrite() {
	log.Println("Listening write to client")
	for {
		select {

		// send message to the client
		case msg := <-c.ch:
			log.Println("Send:", msg)
			websocket.JSON.Send(c.ws, msg)

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenRead method
			return
		}
	}
}

// Listen read request via chanel
func (c *Client) listenRead() {
	log.Println("Listening read from client")
	for {
		select {

		// receive done request
		case <-c.doneCh:
			c.server.Del(c)
			c.doneCh <- true // for listenWrite method
			return

		// read data from websocket connection
		default:
			var msg Message
			err := websocket.JSON.Receive(c.ws, &msg)
			if err == io.EOF {
				c.doneCh <- true
			} else if err != nil {
				c.server.Err(err)
			} else {
				msg.Identity = c.identity
				msg.FromId = c.identity.Id
				msg.CreatedDate = time.Now()
				c.server.SendAll(&msg)
			}
		}
	}
}