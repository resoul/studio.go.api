package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/football.manager.api/internal/config"
	"github.com/football.manager.api/internal/data"
	"github.com/football.manager.api/internal/domain"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	platformdb "github.com/football.manager.api/internal/platform/db"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var (
	adminCreateEmail    string
	adminCreatePassword string
)

var adminCreateCmd = cobra.Command{
	Use:   "admin:create",
	Short: "Create a new admin user",
	Long:  "Creates a new user with role=admin and verified email. User can immediately log in via admin panel.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := config.Init(ctx)

		email := strings.TrimSpace(strings.ToLower(adminCreateEmail))
		if email == "" {
			logrus.Fatal("--email is required")
		}
		if adminCreatePassword == "" {
			logrus.Fatal("--password is required")
		}

		db, err := platformdb.NewDatabase(cfg.DB.DSN)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to connect to database")
		}
		sqlDB, _ := db.DB()
		defer sqlDB.Close()

		userRepo := data.NewUserRepository(db)

		existing, err := userRepo.GetByEmail(ctx, email)
		if err != nil && err != domain.ErrUserNotFound {
			logrus.WithError(err).Fatal("Failed to check existing user")
		}
		if existing != nil {
			if existing.Role == platformauth.RoleAdmin {
				logrus.Fatalf("User %s already exists and is already an admin", email)
			}
			// upgrade existing user to admin
			if err := userRepo.SetRole(ctx, existing.ID, platformauth.RoleAdmin); err != nil {
				logrus.WithError(err).Fatal("Failed to set admin role")
			}
			fmt.Printf("✓ Existing user %s promoted to admin\n", email)
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminCreatePassword), bcrypt.DefaultCost)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to hash password")
		}

		now := time.Now().UTC()
		user := &domain.User{
			UUID:            uuid.New().String(),
			FullName:        "Admin",
			Email:           email,
			PasswordHash:    string(passwordHash),
			Role:            platformauth.RoleAdmin,
			EmailVerifiedAt: &now,
		}

		if err := userRepo.Create(ctx, user); err != nil {
			logrus.WithError(err).Fatal("Failed to create admin user")
		}

		// Ensure role is persisted (Create uses the Role field via mapper, but set explicitly to be safe)
		if err := userRepo.SetRole(ctx, user.ID, platformauth.RoleAdmin); err != nil {
			logrus.WithError(err).Fatal("Failed to set admin role after creation")
		}

		fmt.Printf("✓ Admin user created successfully\n")
		fmt.Printf("  Email: %s\n", email)
		fmt.Printf("  ID:    %d\n", user.ID)
	},
}

func init() {
	adminCreateCmd.Flags().StringVar(&adminCreateEmail, "email", "", "Admin email address (required)")
	adminCreateCmd.Flags().StringVar(&adminCreatePassword, "password", "", "Admin password (required)")
}
