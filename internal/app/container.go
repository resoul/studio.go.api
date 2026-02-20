package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/football.manager.api/internal/config"
	"github.com/football.manager.api/internal/data"
	handler "github.com/football.manager.api/internal/http"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	platformdb "github.com/football.manager.api/internal/platform/db"
	"github.com/football.manager.api/internal/platform/mailer"
	"github.com/football.manager.api/internal/usecase"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Container holds app dependencies built in one place.
type Container struct {
	Config *config.Config
	DB     *gorm.DB

	AuthHandler    *handler.AuthHandler
	UserHandler    *handler.UserHandler
	ManagerHandler *handler.ManagerHandler
	CareerHandler  *handler.CareerHandler

	UserAuthMiddleware  gin.HandlerFunc
	AdminAuthMiddleware gin.HandlerFunc
}

func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.Init(ctx)

	db, err := platformdb.NewDatabase(cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	userRepo := data.NewUserRepository(db)
	managerRepo := data.NewManagerRepository(db)
	careerRepo := data.NewCareerRepository(db)

	emailSender := mailer.NewLogEmailSender()
	if strings.EqualFold(cfg.Mailer.Provider, "smtp") {
		emailSender = mailer.NewSMTPEmailSender(
			cfg.Mailer.Host,
			cfg.Mailer.Port,
			cfg.Mailer.Username,
			cfg.Mailer.Password,
			cfg.Mailer.From,
			cfg.Mailer.LogoPath,
		)
	}

	userTokenManager := platformauth.NewUserTokenManager(cfg.Auth.JWTSecret, time.Duration(cfg.Auth.JWTTTLMinutes)*time.Minute)
	authUC := usecase.NewAuthUseCase(
		userRepo,
		managerRepo,
		userTokenManager,
		emailSender,
		cfg.Mailer.GetAdminEmails(),
	)
	userUC := usecase.NewUserUseCase(userRepo, managerRepo, careerRepo)
	managerUC := usecase.NewManagerUseCase(managerRepo)
	careerUC := usecase.NewCareerUseCase(managerRepo, careerRepo)

	return &Container{
		Config:              cfg,
		DB:                  db,
		AuthHandler:         handler.NewAuthHandler(authUC),
		UserHandler:         handler.NewUserHandler(userUC),
		ManagerHandler:      handler.NewManagerHandler(managerUC),
		CareerHandler:       handler.NewCareerHandler(careerUC),
		UserAuthMiddleware:  platformauth.UserAuthMiddleware(userTokenManager),
		AdminAuthMiddleware: platformauth.AdminAuthMiddleware(userTokenManager, cfg.Auth.GetAdminRoles()),
	}, nil
}

func (c *Container) Close() error {
	if c == nil || c.DB == nil {
		return nil
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
