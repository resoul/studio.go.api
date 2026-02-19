package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	userHandler *UserHandler,
	userAuthMiddleware gin.HandlerFunc,
) {
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/registration", authHandler.Register)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/login", authHandler.Login)
			auth.POST("/reset-password/request", authHandler.RequestResetPassword)
			auth.POST("/reset-password/confirm", authHandler.ConfirmResetPassword)
			auth.GET("/check", userAuthMiddleware, authHandler.CheckAuth)
		}

		// Users routes
		users := v1.Group("/users")
		{
			users.GET("/me", userAuthMiddleware, userHandler.GetMe)
			users.GET("/:id", userHandler.GetUserByID)
		}
	}
}
