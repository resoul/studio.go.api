package cmd

import (
	"github.com/football.manager.api/internal/config"
	"github.com/football.manager.api/internal/data"
	platformdb "github.com/football.manager.api/internal/platform/db"
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

		db, err := platformdb.NewDatabase(cfg.DB.DSN)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to connect to database")
		}

		m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
			{
				ID: "202602192030_create_tables",
				Migrate: func(tx *gorm.DB) error {
					return tx.AutoMigrate(&data.UserModel{}, &data.UserLastLoginModel{}, &data.ManagerModel{})
				},
				Rollback: func(tx *gorm.DB) error {
					return tx.Migrator().DropTable("users", "user_last_logins", "managers")
				},
			},
		})

		if err := m.Migrate(); err != nil {
			logrus.Fatalf("Could not migrate: %v", err)
		}
		logrus.Info("Migration run successfully")
	},
}
