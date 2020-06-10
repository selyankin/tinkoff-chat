package ws

import (
	"chat/pkg/model"
	"time"
)

type Message struct {
	ID 			string 		`json:"id"`
	DestChatID  string      `json:"dest_chat_id"`
	DestChatTitle string    `json:"dest_chat_title"`
	Text        string      `json:"text"`
	FromId      string      `json:"from_id"`
	Identity    *model.User `json:"-"`
	CreatedDate time.Time   `json:"created_data"`
}

func (m *Message) String() string {
	return m.Text + " says " + m.DestChatID
}
