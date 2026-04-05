package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/di"
	"github.com/resoul/studio.go.api/internal/infrastructure/db"
	"github.com/resoul/studio.go.api/internal/service"
	"github.com/resoul/studio.go.api/internal/transport/http/handlers"
	"github.com/resoul/studio.go.api/internal/transport/http/router"
	"github.com/resoul/studio.go.api/internal/worker"
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

	// Repositories
	profileRepo := db.NewProfileRepository(container.DB)
	workspaceRepo := db.NewWorkspaceRepository(container.DB)

	// Services
	profileSvc := service.NewProfileService(profileRepo, container.Storage)
	workspaceSvc := service.NewWorkspaceService(workspaceRepo, container.Storage, container.RabbitMQ)

	// Handlers
	profileHandler := handlers.NewProfileHandler(profileSvc, workspaceSvc)
	workspaceHandler := handlers.NewWorkspaceHandler(workspaceSvc)

	// Start invite worker (async email delivery via RabbitMQ)
	if container.RabbitMQ != nil {
		inviteWorker := worker.NewInviteWorker(container.RabbitMQ, workspaceRepo, container.Mailer)
		go inviteWorker.Start(ctx)
	}

	// Router
	r := router.New(cfg, profileHandler, workspaceHandler)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
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
