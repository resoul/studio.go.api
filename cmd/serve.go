package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/football.manager.api/internal/app"
	handler "github.com/football.manager.api/internal/http"
	"github.com/football.manager.api/internal/platform/httpx"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd, args)
	},
}

func serve(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	container, err := app.NewContainer(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to build app container")
	}
	defer func() {
		if err := container.Close(); err != nil {
			logrus.WithError(err).Error("Error closing database")
		}
	}()

	// Router
	router := gin.Default()
	router.Use(httpx.CORSMiddleware(container.Config.Server.GetCORSAllowedOrigins()))
	handler.RegisterRoutes(
		router,
		container.AuthHandler,
		container.UserHandler,
		container.ManagerHandler,
		container.CareerHandler,
		container.UserAuthMiddleware,
		container.AdminAuthMiddleware,
	)

	// Server
	addr := fmt.Sprintf(":%s", container.Config.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logrus.Infof("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
