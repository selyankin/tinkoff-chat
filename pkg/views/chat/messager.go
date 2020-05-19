package views

import (
	"chat/pkg/model"
	"chat/pkg/utils"
	"chat/pkg/ws"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateChat(c *gin.Context) {
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var createChatRequestPayload struct{
		UserIds []string `json:"user_ids"`
		Name    string `json:"name"`
	}
	err = json.NewDecoder(c.Request.Body).Decode(&createChatRequestPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}

	if len(createChatRequestPayload.UserIds) == 0 || utils.ArrayIn(createChatRequestPayload.UserIds, identity.Id) {
		respErr(c, nil,"failed to create chat")
		return
	}

	mongoClient := c.Keys["mongo"].(*mongo.Client)
	chat, err := model.CreateChat(mongoClient, identity, createChatRequestPayload.UserIds, createChatRequestPayload.Name)
	if err != nil{
		respErr(c, err, "failed to create char")
	}
	c.JSON(200, chat)
	log.Println(identity)
}


func SearchUsers(c *gin.Context) {
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var searchRequestPayload struct{
		Query string `json:"query"`
	}

	err = json.NewDecoder(c.Request.Body).Decode(&searchRequestPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}

	if searchRequestPayload.Query == ""{
		c.JSON(200, []model.User{})
		return
	}

	mongoClient := c.Keys["mongo"].(*mongo.Client)

	users := model.GetUsers(mongoClient, searchRequestPayload.Query, identity)
	fmt.Println(users)
	c.JSON(200, users)

}

func ChatsList(c *gin.Context){
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var searchRequestPayload struct{
		Query string `json:"query"`
	}

	err = json.NewDecoder(c.Request.Body).Decode(&searchRequestPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}

	mongoClient := c.Keys["mongo"].(*mongo.Client)

	chats := model.GetChats(mongoClient, identity, searchRequestPayload.Query)

	c.JSON(200, chats)
}

func ChatInfo(c *gin.Context){
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var searchRequestPayload struct{
		ChatId string `json:"chat_id"`
	}

	err = json.NewDecoder(c.Request.Body).Decode(&searchRequestPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}

	mongoClient := c.Keys["mongo"].(*mongo.Client)

	chat, err := model.ChatInfo(mongoClient, searchRequestPayload.ChatId, identity)
	if err != nil{
		respErr(c, err, "failed to fetch chat")
		return
	}

	c.JSON(200, chat)
}

func SendMessage(c *gin.Context){
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var sendMsgPayload struct{
		ChatId string `json:"chat_id"`
		Message string `json:"message"`
	}

	err = json.NewDecoder(c.Request.Body).Decode(&sendMsgPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}
	mongoClient := c.Keys["mongo"].(*mongo.Client)

	msg := model.SendMessage(mongoClient, sendMsgPayload.ChatId, sendMsgPayload.Message, identity)

	c.JSON(200, msg)
}

func GetMessages(c *gin.Context){
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var getMessagesPayload struct{
		ChatId string `json:"chat_id"`
	}
	err = json.NewDecoder(c.Request.Body).Decode(&getMessagesPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}
	mongoClient := c.Keys["mongo"].(*mongo.Client)
	messages := model.GetMessages(mongoClient, getMessagesPayload.ChatId, identity)

	c.JSON(200, messages)
}

func MarkAsRead(c *gin.Context){
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	var markAsReadPayload struct{
		ChatId string `json:"chat_id"`
		MessageId string `json:"message_id"`
	}
	err = json.NewDecoder(c.Request.Body).Decode(&markAsReadPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}

	mongoClient := c.Keys["mongo"].(*mongo.Client)
	model.MarkAsRead(mongoClient, markAsReadPayload.MessageId, markAsReadPayload.ChatId, identity)

	c.JSON(200, map[string]interface{}{"success": true})
}



//WS

func WebSocket(c *gin.Context){
	wsServer := c.Keys["ws"].(*ws.Server)
	identity, err := model.GetIdentity(c)
	if err != nil {
		c.JSON(400, map[string]interface{}{"success": false, "reason": "unauthorized"})
		return
	}

	handler := wsServer.GetHandler(identity)
	handler.ServeHTTP(c.Writer, c.Request)
}