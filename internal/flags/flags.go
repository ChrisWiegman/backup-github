package flags

import "github.com/spf13/cobra"

// Verbose Set to true for verbose output on a given command.
var Verbose bool

// OutputDir Sets the directory where the backup will write to.
var OutputDir string

// AddVerboseFlag Adds the verbose flag to a cobra command.
func AddVerboseFlag(command *cobra.Command) {
	command.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose (detailed) output.")
}

// AddOutputFlag Adds the output-dir flag to a cobra command.
func AddOutputFlag(command *cobra.Command) {
	command.Flags().StringVarP(&OutputDir, "output-dir", "o", "", "Use relative or absolute path to output backups to.")
}
