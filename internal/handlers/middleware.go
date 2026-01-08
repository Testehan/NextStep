package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		secretKey := os.Getenv("API_SECRET_KEY")
		
		// If no secret key is set in .env, we skip auth (useful for local dev setup)
		if secretKey == "" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer <token>"})
			c.Abort()
			return
		}

		if parts[1] != secretKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid secret key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
