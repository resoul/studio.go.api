package cmd

import (
	"fmt"

	"github.com/football.manager.api/internal/config"
	"github.com/football.manager.api/internal/fixtures"
	platformdb "github.com/football.manager.api/internal/platform/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var seedCmd = cobra.Command{
	Use:   "seed",
	Short: "Seed database with initial fixture data",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cfg := config.Init(ctx)

		db, err := platformdb.NewDatabase(cfg.DB.DSN)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to connect to database")
		}
		sqlDB, _ := db.DB()
		defer sqlDB.Close()

		n, err := fixtures.SeedCountries(ctx, db)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to seed countries")
		}
		fmt.Printf("✓ Countries seeded (%d rows affected)\n", n)
	},
}
