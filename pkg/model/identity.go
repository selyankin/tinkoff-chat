package model

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Session struct {
	Login       string    `json:"login"`
	Token       string    `json:"token"`
	ActiveUntil time.Time `json:"active_until"`
}

type User struct {
	Id       string `json:"_id" bson:"_id,omitempty"`
	Login    string `json:"login" bson:"login"`
	Password string `json:"password" bson:"-"`
	PHash    string `json:"p_hash"`
	Email    string `json:"email" bson:"email"`
	Name     string `json:"name"  bson:"name"`
}

func GetIdentity(c *gin.Context) (*User, error) {
	token := c.Request.Header.Get("X-Auth-Token")

	mongoClient := c.Keys["mongo"].(*mongo.Client)

	sessionsCollection := mongoClient.Database("chat").Collection("sessions")
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	sessionResp := sessionsCollection.FindOne(ctx, bson.M{"token": token})
	if sessionResp.Err() != nil {
		return nil, errors.New("No such session")
	}

	var s Session
	err := sessionResp.Decode(&s)
	if err != nil {
		return nil, errors.New("Broken session")
	}
	if s.ActiveUntil.Before(time.Now()) {
		return nil, errors.New("Session expired")
	}

	usersCollection := mongoClient.Database("chat").Collection("users")

	userResp := usersCollection.FindOne(ctx, bson.M{"login": s.Login})
	if userResp.Err() != nil {
		return nil, errors.New("No such user")
	}

	var u User
	err = userResp.Decode(&u)
	if err != nil {
		return nil, errors.New("Broken user")
	}

	return &u, nil
}
