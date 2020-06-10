package model

import (
	"chat/pkg/utils"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Chat struct {
	Id              string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name            string    `json:"name" bson:"name"`
	UserIds         []string  `json:"user_ids" bson:"user_ids"`
	MessageIds      []string  `json:"message_ids" bson:"message_ids"`
	Owner           string    `json:"owner" bson:"owner"`
	LastMessageDate time.Time `json:"last_message_date" bson:"last_message_date"`
	LastMessageText string    `json:"last_message_text" bson:"last_message_text"`
}

func CreateChat(mongoClient *mongo.Client, identity *User, userIds []string, name string) (*Chat, error) {
	chatsCollection := mongoClient.Database("chat").Collection("chats")
	userIds = append(userIds, identity.Id)
	c := Chat{
		Name:       name,
		UserIds:    utils.Unique(userIds),
		MessageIds: []string{},
		Owner:      identity.Id,
		LastMessageDate: time.Now(),
	}
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	result, err := chatsCollection.InsertOne(ctx, c)

	if err != nil {
		return nil, err
	}
	fmt.Println(result.InsertedID.(primitive.ObjectID).Hex())
	return &Chat{
		Id:         result.InsertedID.(primitive.ObjectID).Hex(),
		Name:       c.Name,
		UserIds:    c.UserIds,
		MessageIds: c.MessageIds,
		Owner:      c.Owner,
		LastMessageDate: c.LastMessageDate,
	}, nil

}

func ChatInfo(mongoClient *mongo.Client, chatID string, identity *User) (*Chat, error) {
	chatsCollection := mongoClient.Database("chat").Collection("chats")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	var chat Chat

	//TODO: Ограничить вывод чатов по доступу
	//userId, err := primitive.ObjectIDFromHex(identity.Id)
	//if err != nil{
	//	return nil, err
	//}
	//"user_ids": userId

	chatId, err := primitive.ObjectIDFromHex(chatID)
	if err != nil {
		return nil, err
	}

	searchResult := chatsCollection.FindOne(ctx, bson.M{"_id": chatId})

	err = searchResult.Decode(&chat)
	fmt.Println(chat)
	if err != nil {
		return nil, errors.New("failed to get response from db")
	}
	return &chat, nil
}

type userSearchResult struct {
	Id    string `json:"id" bson:"_id"`
	Login string `json:"login" bson:"login"`
	Name  string `json:"name" bson:"name"`
}

func GetUsers(mongoClient *mongo.Client, searchString string, identity *User) []userSearchResult {
	usersCollection := mongoClient.Database("chat").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	options := options.Find()
	options.SetLimit(5)

	searchResult, err := usersCollection.Find(ctx, bson.M{"login": bson.M{"$regex": fmt.Sprintf("^%s", searchString), "$options": "i"}}, options)
	if err != nil {
		log.Error(err)
		return []userSearchResult{}
	}

	var users []userSearchResult
	err = searchResult.All(ctx, &users)
	if err != nil {
		log.Error(err)
		return []userSearchResult{}
	}

	var result []userSearchResult

	for _, user := range users {
		if user.Id == identity.Id {
			continue
		}
		result = append(result, user)
	}

	if result == nil || len(result) == 0 {
		return []userSearchResult{}
	}

	return result
}

func GetChats(client *mongo.Client, identity *User, query string) []Chat {
	chatsCollection := client.Database("chat").Collection("chats")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	searchResult, err := chatsCollection.Find(ctx, bson.M{"user_ids": identity.Id, "name": bson.M{"$regex": fmt.Sprintf("^%s", query), "$options": "i"}})

	if err != nil {
		log.Error(err)
		return []Chat{}
	}

	var chats []Chat

	err = searchResult.All(ctx, &chats)
	if err != nil {
		log.Error(err)
		return []Chat{}
	}
	if chats == nil {
		return []Chat{}
	}
	return chats
}

type Message struct {
	Id     string    `json:"id,omitempty" bson:"_id,omitempty"`
	ChatId string    `json:"chat_id" bson:"chat_id"`
	Owner  string    `json:"owner" bson:"owner"`
	Text   string    `json:"text" bson:"text"`
	Date   time.Time `json:"date" bson:"date"`
	Viewed bool      `json:"viewed" bson:"viewed"`
}

func getChat(client *mongo.Client, chatId string) *Chat {
	chatsCollection := client.Database("chat").Collection("chats")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	objId, err := primitive.ObjectIDFromHex(chatId)
	if err != nil {
		return nil
	}
	searchResult, err := chatsCollection.Find(ctx, bson.M{"_id": objId})
	if err != nil {
		log.Error(err)
		return nil
	}

	var chats []Chat
	err = searchResult.All(ctx, &chats)
	fmt.Println(chats)
	if err != nil {
		log.Error(err)
		return nil
	}
	if chats == nil || len(chats) == 0 {
		return nil
	}

	return &chats[0]
}

func SendMessage(client *mongo.Client, chatId string, message string, identity *User) Message {
	messagesCollection := client.Database("chat").Collection("messages")

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	chat := getChat(client, chatId)

	if chat == nil || !utils.ArrayIn(chat.UserIds, identity.Id) {
		return Message{}
	}

	msg := Message{
		ChatId: chatId,
		Text:   message,
		Owner:  identity.Id,
		Date:   time.Now(),
	}

	result, err := messagesCollection.InsertOne(ctx, msg)

	if err != nil {
		return Message{}
	}

	return Message{
		Id:     result.InsertedID.(primitive.ObjectID).Hex(),
		ChatId: msg.ChatId,
		Text:   msg.Text,
		Owner:  msg.Owner,
		Date:   msg.Date,
	}
}

func GetMessages(client *mongo.Client, chatId string, identity *User) []Message {
	fmt.Println(identity)
	messagesCollection := client.Database("chat").Collection("messages")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	chat := getChat(client, chatId)
	fmt.Println(chat)
	if chat == nil || !utils.ArrayIn(chat.UserIds, identity.Id) {
		return []Message{}
	}

	searchResult, err := messagesCollection.Find(ctx, bson.M{"chat_id": chat.Id})
	if err != nil {
		log.Error(err)
		return []Message{}
	}

	var messages []Message
	err = searchResult.All(ctx, &messages)
	if err != nil {
		log.Error(err)
		return []Message{}
	}
	if messages == nil || len(messages) == 0 {
		return []Message{}
	}

	return messages

}

func MarkAsRead(client *mongo.Client, messageId string, chatId string, identity *User) {
	messagesCollection := client.Database("chat").Collection("messages")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	chat := getChat(client, chatId)
	if chat == nil || !utils.ArrayIn(chat.UserIds, identity.Id) {
		return
	}
	id, _ := primitive.ObjectIDFromHex(messageId)
	_, _ = messagesCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"viewed": true}})
}

func UpdateChatLastMessage(client *mongo.Client, chatId string, lastMessageDate time.Time, lastMessageText string) {
	messagesCollection := client.Database("chat").Collection("chats")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	id, _ := primitive.ObjectIDFromHex(chatId)
	_, _ = messagesCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"last_message_date": lastMessageDate, "last_message_text": lastMessageText}})
}
