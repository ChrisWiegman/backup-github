package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string
var Timestamp string

type VersionInfo struct {
	Version, Timestamp string
}

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version", //nolint:goconst //Additional usages are in testing.
		Short: "Displays version information for the Backup GitHub script.",
		Run: func(cmd *cobra.Command, args []string) { //nolint:revive //Parameters are required by Cobra.
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Build Time: %s\n", Timestamp)
		},
		Args: cobra.NoArgs,
	}

	return cmd
}
