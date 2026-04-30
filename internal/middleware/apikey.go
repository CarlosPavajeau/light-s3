package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKey(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader("X-API-Key")
		if provided == "" {
			provided = c.Query("api_key")
		}
		if provided != key {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
