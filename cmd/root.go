package cmd

import (
	"sync"

	"github.com/spf13/cobra"
)

var (
	rootCmd = cobra.Command{
		Use: "score",
	}
	mainWG *sync.WaitGroup
)

func RootCommand(wg *sync.WaitGroup) *cobra.Command {
	mainWG = wg
	rootCmd.AddCommand(&serveCmd, &migrateCmd)
	return &rootCmd
}
