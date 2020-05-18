package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "X-Auth-Token, Content-Type, Accept, Origin")
		if c.Request.Method == http.MethodOptions{
			c.JSON(204, nil)
		}
	}
}
