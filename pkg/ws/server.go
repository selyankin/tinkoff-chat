package ws

import (
	"chat/pkg/model"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/websocket"
	"log"
)

// Chat server.
type Server struct {
	mongoClient *mongo.Client
	messages  []*Message
	clients   map[string]*Client
	addCh     chan *Client
	delCh     chan *Client
	sendAllCh chan *Message
	doneCh    chan bool
	errCh     chan error
}

// Create new chat server.
func NewServer(mongoClient *mongo.Client) *Server {
	messages := []*Message{}
	clients := make(map[string]*Client)
	addCh := make(chan *Client)
	delCh := make(chan *Client)
	sendAllCh := make(chan *Message)
	doneCh := make(chan bool)
	errCh := make(chan error)

	return &Server{
		mongoClient,
		messages,
		clients,
		addCh,
		delCh,
		sendAllCh,
		doneCh,
		errCh,
	}
}

func (s *Server) Add(c *Client) {
	s.addCh <- c
}

func (s *Server) Del(c *Client) {
	s.delCh <- c
}

func (s *Server) SendAll(msg *Message) {
	s.sendAllCh <- msg
}

func (s *Server) Done() {
	s.doneCh <- true
}

func (s *Server) Err(err error) {
	s.errCh <- err
}

func (s *Server) SendMessage(msg *Message) {
	chat, err := model.ChatInfo(s.mongoClient, msg.DestChatID, msg.Identity)

	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(*msg)
	fmt.Println(chat.UserIds)

	var destUsers []*Client

	for _, userID := range chat.UserIds{
		if _, ok := s.clients[userID]; ok {
			fmt.Println("Sending to ", userID)
			destUsers = append(destUsers, s.clients[userID])
		}
	}
	sendedMsg := model.SendMessage(s.mongoClient,msg.DestChatID,msg.Text,msg.Identity)
	model.UpdateChatLastMessage(s.mongoClient,msg.DestChatID, msg.CreatedDate, msg.Text)

	msg.ID = sendedMsg.Id
	msg.DestChatTitle = chat.Name
	for _, c := range destUsers {
		c.Write(msg)
	}

}

// Listen and serve.
// It serves client connection and broadcast request.
func (s *Server) Listen() {
	for {
		select {

		// Add new a client
		case c := <-s.addCh:
			log.Println("Added new client")
			s.clients[c.identity.Id] = c
			log.Println("Now", len(s.clients), "clients connected.")

		// del a client
		case c := <-s.delCh:
			log.Println("Delete client")
			delete(s.clients, c.identity.Id)

		// TODO: Send to specific clients
		case msg := <-s.sendAllCh:
			log.Println("Send all:", msg)
			//s.messages = append(s.messages, msg)
			s.SendMessage(msg)

		case err := <-s.errCh:
			log.Println("Error:", err.Error())

		case <-s.doneCh:
			return
		}
	}
}

func (s *Server) GetHandler(identity *model.User) websocket.Handler{
	log.Println("Listening server...")

	// websocket handler
	onConnected := func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				s.errCh <- err
			}
		}()

		client := NewClient(ws, s, identity)
		s.Add(client)
		client.Listen()
	}
	return websocket.Handler(onConnected)
}
