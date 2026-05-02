package cmd

import (
	"fmt"
	"os"

	"github.com/ChrisWiegman/backup-github/internal/backup"

	"github.com/spf13/cobra"
)

var flagVersion bool
var Version, Timestamp string

type VersionInfo struct {
	Version, Timestamp string
}

func Execute() {
	// Set up the cobra command
	cmd := &cobra.Command{
		Use:   "backup-github",
		Short: "Backup GitHub is a simple script to backup all your GitHub repos.",
		Args:  cobra.MaximumNArgs(1),
		Run:   runCommand,
	}

	cmd.PersistentFlags().BoolVarP(&flagVersion, "version", "v", false, "Display version information for the app.")

	// Execute anything we need to
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCommand(cmd *cobra.Command, args []string) {
	if cmd.Flags().Lookup("version").Value.String() == "true" {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", Timestamp)

		os.Exit(0)
	}

	backup.ExecuteBackup()
}
