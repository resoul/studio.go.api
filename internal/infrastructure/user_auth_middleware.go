package infrastructure

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func UserAuthMiddleware(tokenManager *UserTokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid authorization format. Use: Bearer <token>"})
			c.Abort()
			return
		}

		userID, err := tokenManager.Parse(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		return 0, false
	}

	return userID, true
}
