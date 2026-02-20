package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/football.manager.api/internal/data"
	handler "github.com/football.manager.api/internal/http"
	"github.com/football.manager.api/internal/infrastructure"
	"github.com/football.manager.api/internal/usecase"
	"github.com/football.manager.api/pkg/config"
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
	cfg := config.Init(ctx)

	// Database
	db, err := infrastructure.NewDatabase(cfg.DB.DSN)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	sqlDB, _ := db.DB()
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logrus.WithError(err).Error("Error closing database")
		}
	}()

	userRepo := data.NewUserRepository(db)

	emailSender := infrastructure.NewLogEmailSender()
	if strings.EqualFold(cfg.Mailer.Provider, "smtp") {
		emailSender = infrastructure.NewSMTPEmailSender(
			cfg.Mailer.Host,
			cfg.Mailer.Port,
			cfg.Mailer.Username,
			cfg.Mailer.Password,
			cfg.Mailer.From,
			cfg.Mailer.LogoPath,
		)
	}

	userTokenManager := infrastructure.NewUserTokenManager(cfg.Auth.JWTSecret, time.Duration(cfg.Auth.JWTTTLMinutes)*time.Minute)
	authUC := usecase.NewAuthUseCase(userRepo, userTokenManager, emailSender, cfg.Mailer.GetAdminEmails())
	userUC := usecase.NewUserUseCase(userRepo)

	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)

	userAuthMiddleware := infrastructure.UserAuthMiddleware(userTokenManager)

	// Router
	router := gin.Default()
	router.Use(handler.CORSMiddleware(cfg.Server.GetCORSAllowedOrigins()))
	handler.RegisterRoutes(router, authHandler, userHandler, userAuthMiddleware)

	// Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
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
