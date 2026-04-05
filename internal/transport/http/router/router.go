package router

import (
	"github.com/gin-gonic/gin"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/transport/http/handlers"
	"github.com/resoul/studio.go.api/internal/transport/http/middleware"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

func New(cfg *config.Config, profileHandler *handlers.ProfileHandler, workspaceHandler *handlers.WorkspaceHandler) *gin.Engine {
	r := gin.Default()
	r.Use(utils.CORSMiddleware(cfg.Server.GetCORSAllowedOrigins()))

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			utils.RespondOK(c, gin.H{"status": "ok"})
		})

		api.GET("/workspaces/invites/:token/preview", workspaceHandler.GetInvitePreview)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/user/me", profileHandler.GetMe)
			protected.PATCH("/user/profile", profileHandler.UpdateProfile)

			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", workspaceHandler.List)
				workspaces.POST("", workspaceHandler.Create)
				workspaces.POST("/invites/:token/accept", workspaceHandler.AcceptInvite)
				workspaces.POST("/:id/invites", workspaceHandler.CreateInvite)
				workspaces.PATCH("/:id", workspaceHandler.Update)

				workspaces.GET("/current", workspaceHandler.GetCurrent)
				workspaces.POST("/current/:id", workspaceHandler.SetCurrent)

				workspaces.GET("/config", workspaceHandler.GetConfig)
				workspaces.PATCH("/config/:id", workspaceHandler.UpdateConfig)
			}
		}
	}

	return r
}
