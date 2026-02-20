package cmd

import (
	"github.com/football.manager.api/internal/data"
	"github.com/football.manager.api/internal/infrastructure"
	"github.com/football.manager.api/pkg/config"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateCmd = cobra.Command{
	Use:   "migrate",
	Short: "Run versioned database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := config.Init(ctx)

		db, err := infrastructure.NewDatabase(cfg.DB.DSN)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to connect to database")
		}

		m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
			{
				ID: "202602192030_create_users_table",
				Migrate: func(tx *gorm.DB) error {
					return tx.AutoMigrate(&data.UserModel{})
				},
				Rollback: func(tx *gorm.DB) error {
					return tx.Migrator().DropTable("users")
				},
			},
			{
				ID: "202602201200_add_user_client_metadata",
				Migrate: func(tx *gorm.DB) error {
					cols := []string{
						"RegistrationIP",
						"RegistrationUserAgent",
						"LoginCount",
					}
					for _, col := range cols {
						if tx.Migrator().HasColumn(&data.UserModel{}, col) {
							continue
						}
						if err := tx.Migrator().AddColumn(&data.UserModel{}, col); err != nil {
							return err
						}
					}
					return nil
				},
				Rollback: func(tx *gorm.DB) error {
					cols := []string{
						"RegistrationIP",
						"RegistrationUserAgent",
						"LoginCount",
					}
					for _, col := range cols {
						if !tx.Migrator().HasColumn(&data.UserModel{}, col) {
							continue
						}
						if err := tx.Migrator().DropColumn(&data.UserModel{}, col); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				ID: "202602201430_move_last_login_to_separate_table",
				Migrate: func(tx *gorm.DB) error {
					if err := tx.AutoMigrate(&data.UserLastLoginModel{}); err != nil {
						return err
					}

					statements := []string{
						"ALTER TABLE users DROP COLUMN IF EXISTS last_login_at",
						"ALTER TABLE users DROP COLUMN IF EXISTS last_login_ip",
						"ALTER TABLE users DROP COLUMN IF EXISTS last_login_user_agent",
					}
					for _, stmt := range statements {
						if err := tx.Exec(stmt).Error; err != nil {
							return err
						}
					}

					return nil
				},
				Rollback: func(tx *gorm.DB) error {
					statements := []string{
						"ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP NULL",
						"ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip VARCHAR(45) NOT NULL DEFAULT ''",
						"ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_user_agent VARCHAR(512) NOT NULL DEFAULT ''",
					}
					for _, stmt := range statements {
						if err := tx.Exec(stmt).Error; err != nil {
							return err
						}
					}

					return tx.Migrator().DropTable(&data.UserLastLoginModel{})
				},
			},
		})

		if err := m.Migrate(); err != nil {
			logrus.Fatalf("Could not migrate: %v", err)
		}
		logrus.Info("Migration run successfully")
	},
}
