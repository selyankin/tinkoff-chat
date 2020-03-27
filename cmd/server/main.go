package main

import (
	"chat/pkg/middleware"
	views "chat/pkg/views/chat"
	"context"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_DSN")))

	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("ddsd")
	r := gin.Default()
	r.Use(middleware.ChatContext(client))
	r.POST("/login", views.Login)
	r.POST("/register", views.Register)

	r.POST("/create_chat", views.CreateChat)

	r.POST("/search_users", views.SearchUsers)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
