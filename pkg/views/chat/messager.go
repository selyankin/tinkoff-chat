package views

import (
	"chat/pkg/model"
	"encoding/json"
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

	mongoClient := c.Keys["mongo"].(*mongo.Client)
	chat, err := model.CreateChat(mongoClient, identity, createChatRequestPayload.UserIds, createChatRequestPayload.Name)
	if err != nil{
		respErr(c, err, "failed to create char")
	}
	c.JSON(200, chat)
	log.Println(identity)
}


func SearchUsers(c *gin.Context) {
	_, err := model.GetIdentity(c)
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

	users := model.GetUsers(mongoClient, searchRequestPayload.Query)
	log.Println(users)
	c.JSON(200, users)

}