package flags

import "github.com/spf13/cobra"

// Verbose Set to true for verbose output on a given command.
var Verbose bool

// AddVerboseFlag Adds the verbose flag to a cobra command.
func AddVerboseFlag(command *cobra.Command) {
	command.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose (detailed) output.")
}
