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
		})

		if err := m.Migrate(); err != nil {
			logrus.Fatalf("Could not migrate: %v", err)
		}
		logrus.Info("Migration run successfully")
	},
}
