package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/di"
	"github.com/resoul/studio.go.api/internal/infrastructure/db"
	"github.com/resoul/studio.go.api/internal/service"
	"github.com/resoul/studio.go.api/internal/transport/http/handlers"
	"github.com/resoul/studio.go.api/internal/transport/http/middleware"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd)
	},
}

func serve(cmd *cobra.Command) {
	ctx := cmd.Context()
	cfg := config.Init(ctx)

	container, err := di.NewContainer(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize container")
	}
	defer container.Close()

	profileRepo := db.NewProfileRepository(container.DB)
	workspaceRepo := db.NewWorkspaceRepository(container.DB)

	profileSvc := service.NewProfileService(profileRepo, container.Storage)
	workspaceSvc := service.NewWorkspaceService(workspaceRepo, container.Storage)

	profileHandler := handlers.NewProfileHandler(profileSvc, workspaceSvc)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceSvc)

	router := gin.Default()
	router.Use(utils.CORSMiddleware(cfg.Server.GetCORSAllowedOrigins()))

	api := router.Group("/api/v1")
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
			}
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logrus.Infof("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	logrus.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}
