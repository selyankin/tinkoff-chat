package middleware

import (
	"chat/pkg/ws"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ChatContext(mongoClient *mongo.Client, wsServer *ws.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("mongo", mongoClient)
		c.Set("ws", wsServer)
		c.Next()
	}
}
