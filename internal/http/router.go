package http

import (
	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.Engine,
	authHandler *AuthHandler,
	userHandler *UserHandler,
	managerHandler *ManagerHandler,
	careerHandler *CareerHandler,
	userAuthMiddleware gin.HandlerFunc,
	adminAuthMiddleware gin.HandlerFunc,
) {
	router.GET("/", func(c *gin.Context) {
		httpx.RespondOK(c, gin.H{
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
			auth.POST("/verify-email/resend", authHandler.ResendVerification)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/reset-password/request", authHandler.RequestResetPassword)
			auth.POST("/reset-password/confirm", authHandler.ConfirmResetPassword)
			auth.GET("/check", userAuthMiddleware, authHandler.CheckAuth)
		}

		admin := v1.Group("/admin")
		{
			adminAuth := admin.Group("/auth")
			{
				adminAuth.POST("/login", authHandler.AdminLogin)
				adminAuth.GET("/check", adminAuthMiddleware, authHandler.CheckAuth)
			}
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", adminAuthMiddleware, userHandler.ListUsers)
			}
		}

		// Users routes
		users := v1.Group("/users")
		{
			users.GET("/me", userAuthMiddleware, userHandler.GetMe)
			users.GET("/:id", adminAuthMiddleware, userHandler.GetUserByID)
		}

		managers := v1.Group("/managers")
		{
			managers.GET("/me", userAuthMiddleware, managerHandler.GetMe)
			managers.POST("/me", userAuthMiddleware, managerHandler.CreateMe)
		}

		careers := v1.Group("/careers")
		{
			careers.GET("/me", userAuthMiddleware, careerHandler.ListMe)
			careers.POST("/me", userAuthMiddleware, careerHandler.CreateMe)
		}
	}
}
