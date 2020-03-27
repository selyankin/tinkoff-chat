package middleware

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ChatContext(mongoClient *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("mongo", mongoClient)
		c.Next()
	}
}
