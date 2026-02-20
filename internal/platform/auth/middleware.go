package auth

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(tokenManager *UserTokenManager, allowedRoles ...string) gin.HandlerFunc {
	normalizedRoles := make([]string, 0, len(allowedRoles))
	for _, role := range allowedRoles {
		role = strings.TrimSpace(strings.ToLower(role))
		if role != "" {
			normalizedRoles = append(normalizedRoles, role)
		}
	}

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

		token := parts[1]

		claims, err := tokenManager.Parse(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid or expired token"})
			c.Abort()
			return
		}

		if len(normalizedRoles) > 0 && !slices.Contains(normalizedRoles, claims.Role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Insufficient role permissions"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func UserAuthMiddleware(tokenManager *UserTokenManager) gin.HandlerFunc {
	return AuthMiddleware(tokenManager, RoleUser, RoleAdmin)
}

func AdminAuthMiddleware(tokenManager *UserTokenManager, adminRoles []string) gin.HandlerFunc {
	if len(adminRoles) == 0 {
		return AuthMiddleware(tokenManager, RoleAdmin)
	}
	return AuthMiddleware(tokenManager, adminRoles...)
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

func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	roleValue, exists := c.Get("user_role")
	if !exists {
		return "", false
	}

	role, ok := roleValue.(string)
	if !ok || role == "" {
		return "", false
	}

	return role, true
}
