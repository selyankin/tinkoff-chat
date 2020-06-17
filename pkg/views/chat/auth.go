package views

import (
	"chat/pkg/model"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	log "github.com/sirupsen/logrus"
	"time"
)

// проверка на то что сессия не протухла во всех хэндлерах

func Login(c *gin.Context) {
	mongoClient := c.Keys["mongo"].(*mongo.Client)

	var reqPayload struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(c.Request.Body).Decode(&reqPayload)

	if err != nil {
		respErr(c, err, "no such user")
		return
	}
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	usersCollection := mongoClient.Database("chat").Collection("users")
	resp := usersCollection.FindOne(ctx, bson.M{"login": reqPayload.Login})

	if resp.Err() != nil {
		respErr(c, err, "no such user")
		return
	}
	var user model.User
	err = resp.Decode(&user)

	if !checkPasswordHash(reqPayload.Password, user.PHash) {
		respErr(c, nil, "incorrect password")
		return
	}

	if err != nil {
		respErr(c, err, "")
	}

	sessionCollection := mongoClient.Database("chat").Collection("sessions")
	sessionResp := sessionCollection.FindOne(ctx, bson.M{"login": user.Login})

	if sessionResp.Err() == nil {
		// если сессия существует то она существует
		var s model.Session
		err = sessionResp.Decode(&s)
		if err != nil {
			respErr(c, err, "")
			return
		}
		if s.ActiveUntil.After(time.Now()) {
			c.JSON(200, map[string]interface{}{"success": true, "token": s.Token, "login": user.Login, "id": user.Id})
			return
		}
	}

	// если сессия не существует
	s := model.Session{
		Login:       user.Login,
		Token:       randToken(),
		ActiveUntil: time.Now().Add(time.Hour * 24),
	}

	bsonReq, err := bson.Marshal(s)
	if err != nil {
		respErr(c, err, "")
		return
	}
	_, err = sessionCollection.InsertOne(ctx, bsonReq)
	if err != nil {
		respErr(c, err, "")
	}

	c.JSON(200, map[string]interface{}{"success": true, "token": s.Token, "login": user.Login, "id": user.Id})
}

func Register(c *gin.Context) {
	// todo: отправлять подтверждение на почту
	mongoClient := c.Keys["mongo"].(*mongo.Client)

	var reqPayload model.User
	err := json.NewDecoder(c.Request.Body).Decode(&reqPayload)

	if err != nil {
		respErr(c, err, "")
		return
	}
	reqPayload.PHash, _ = hashPassword(reqPayload.Password)

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

	bsonReq, err := bson.Marshal(reqPayload)
	if err != nil {
		respErr(c, err, "")
		return
	}

	res, err := mongoClient.Database("chat").Collection("users").InsertOne(ctx, bsonReq)
	logrus.Info(res)
	if err != nil {
		respErr(c, err, "")
		return
	}
	c.JSON(200, map[string]interface{}{"success": true})
}

func respErr(c *gin.Context, err error, msg string) {
	log.Error(err)
	c.JSON(500, map[string]interface{}{"error": err, "success": false, "msg": msg})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func IsValidToken(c *gin.Context){

	var validTokenPayload struct{
		Token string `json:"token"`
	}
	err := json.NewDecoder(c.Request.Body).Decode(&validTokenPayload)
	if err != nil{
		respErr(c, err, "failed to read body")
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	mongoClient := c.Keys["mongo"].(*mongo.Client)

	sessionCollection := mongoClient.Database("chat").Collection("sessions")
	sessionResp := sessionCollection.FindOne(ctx, bson.M{"token": validTokenPayload.Token})

	if sessionResp.Err() == nil {
		var s model.Session
		err := sessionResp.Decode(&s)
		if err != nil {
			respErr(c, err, "")
			return
		}
		if s.ActiveUntil.After(time.Now()) {
			c.JSON(200, map[string]interface{}{"success": true})
			return
		}
	}

	respErr(c, nil, "")
}
