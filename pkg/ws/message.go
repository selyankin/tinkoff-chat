package ws

import "chat/pkg/model"

type Message struct {
	DestChatID   string `json:"dest_chat_id"`
	Text string `json:"text"`
	FromId string `json:"from_id"`
	Identity *model.User `json:"-"`
}

func (m *Message) String() string {
	return m.Text + " says " + m.DestChatID
}
