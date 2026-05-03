package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ChrisWiegman/backup-github/internal/client"
)

func logoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout", //nolint:goconst //Additional usages are in testing.
		Short: "Logs the user out of GitHub. You'll need to login again when next running the app.",
		RunE: func(cmd *cobra.Command, args []string) error { //nolint:revive //Parameters are required by Cobra.
			err := client.LogoutGitHub()
			if err == nil {
				fmt.Println("You have successfully logged out of Backup GitHub.")
			}

			return err
		},
		Args: cobra.NoArgs,
	}

	return cmd
}
