package cmd

import (
	"os"

	"github.com/ChrisWiegman/backup-github/internal/backup"
	"github.com/ChrisWiegman/backup-github/internal/flags"

	"github.com/spf13/cobra"
)

func Execute() {
	cmd := rootCommand()

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func rootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup-github", //nolint:goconst //Additional usages are in testing.
		Short: "Backup GitHub is a simple script to backup all your GitHub repos.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCommand,
	}

	flags.AddVerboseFlag(cmd)
	flags.AddOutputFlag(cmd)

	cmd.AddCommand(
		versionCommand(),
		logoutCommand(),
	)

	return cmd
}

func runCommand(cmd *cobra.Command, args []string) error { //nolint:revive //Passing args is required by Cobra.
	return backup.ExecuteBackup(flags.Verbose, flags.OutputDir, cmd.Flags().Lookup("output-dir").Changed)
}
