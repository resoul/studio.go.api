package router

import (
	"github.com/gin-gonic/gin"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/transport/http/handlers"
	"github.com/resoul/studio.go.api/internal/transport/http/middleware"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

func New(cfg *config.Config, profileHandler *handlers.ProfileHandler, workspaceHandler *handlers.WorkspaceHandler, wsHandler *handlers.WSHandler, chatHandler *handlers.ChatHandler) *gin.Engine {
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
			protected.GET("/ws", wsHandler.HandleWS)

			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", workspaceHandler.List)
				workspaces.POST("", workspaceHandler.Create)
				workspaces.POST("/invites/:token/accept", workspaceHandler.AcceptInvite)
				workspaces.POST("/:id/invites", workspaceHandler.CreateInvite)
				workspaces.GET("/:id/invites", workspaceHandler.ListInvites)
				workspaces.POST("/:id/invites/resend", workspaceHandler.ResendInvite)
				workspaces.DELETE("/:id/invites/:email", workspaceHandler.RevokeInvite)
				workspaces.PATCH("/:id", workspaceHandler.Update)

				workspaces.GET("/:id/members", workspaceHandler.ListMembers)
				workspaces.DELETE("/:id/members/:user_id", workspaceHandler.RemoveMember)

				workspaces.GET("/current", workspaceHandler.GetCurrent)
				workspaces.POST("/current/:id", workspaceHandler.SetCurrent)

				workspaces.GET("/config", workspaceHandler.GetConfig)
				workspaces.PATCH("/config/:id", workspaceHandler.UpdateConfig)

				chat := workspaces.Group("/:id/chat")
				{
					chat.GET("/channels", chatHandler.ListChannels)
					chat.POST("/channels", chatHandler.CreateChannel)
					chat.GET("/conversations", chatHandler.ListConversations)
					chat.POST("/conversations", chatHandler.GetOrCreateConversation)
					chat.GET("/messages/:chat_id", chatHandler.GetMessages) // changed from :id to :chat_id to avoid confusion with workspace :id
					chat.GET("/messages/:chat_id/thread/:message_id", chatHandler.GetThreadMessages)
					chat.POST("/messages/:chat_id", chatHandler.SendMessage)
					chat.POST("/messages/:chat_id/read", chatHandler.MarkRead)
					chat.POST("/reactions/:message_id", chatHandler.ToggleReaction)
				}
			}
		}
	}

	return r
}
