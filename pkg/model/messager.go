package model

import (
	"chat/pkg/utils"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)



type Chat struct {

	Name       string   `json:"name" bson:"name"`
	UserIds    []string `json:"user_ids" bson:"user_ids"`
	MessageIds []string `json:"message_ids" bson:"message_ids"`
	Owner      string   `json:"owner" bson:"owner"`
}

type RWChat struct {
	Id string `json:"id" bson:"_id"`
	Chat
}

func CreateChat(mongoClient *mongo.Client, identity *RWUser, userIds []string, name string) (*RWChat, error) {
	chatsCollection := mongoClient.Database("chat").Collection("chats")
	userIds = append(userIds, identity.Id)
	c := Chat{
		Name:       name,
		UserIds:    utils.Unique(userIds),
		MessageIds: []string{},
		Owner:      identity.Id,
	}
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	result, err := chatsCollection.InsertOne(ctx, c)

	if err != nil{
		return nil, err
	}

	return &RWChat{
		Id:   result.InsertedID.(primitive.ObjectID).String(),
		Chat: c,
	}, nil

}


type userSearchResult struct {
	Id string `json:"id" bson:"_id"`
	Login string `json:"login" bson:"login"`
	Name  string `json:"name" bson:"name"`
}

func GetUsers(mongoClient *mongo.Client, searchString string) []userSearchResult {
	usersCollection := mongoClient.Database("chat").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	searchResult, err := usersCollection.Find(ctx, bson.M{"login": bson.M{"$regex": fmt.Sprintf("^%s", searchString), "$options": "i"}})
	if err != nil{
		log.Error(err)
		return []userSearchResult{}
	}

	var users []userSearchResult
	err = searchResult.All(ctx, &users)
	if err != nil{
		log.Error(err)
		return []userSearchResult{}
	}
	if users == nil{
		return []userSearchResult{}
	}
	return users
}