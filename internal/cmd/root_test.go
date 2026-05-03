package cmd

import (
	"testing"
)

func TestRootCommand_Structure(t *testing.T) {
	cmd := rootCommand()
	if cmd.Use != "backup-github" {
		t.Errorf("Use = %q, want 'backup-github'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short description must not be empty")
	}
}

func TestRootCommand_Subcommands(t *testing.T) {
	cmd := rootCommand()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{"version", "logout"} {
		if !names[want] {
			t.Errorf("rootCommand missing %q subcommand", want)
		}
	}
}

func TestRootCommand_OutputDirFlag(t *testing.T) {
	cmd := rootCommand()
	flag := cmd.Flags().Lookup("output-dir")
	if flag == nil {
		t.Fatal("rootCommand missing --output-dir flag")
	}
	if flag.Shorthand != "o" {
		t.Errorf("output-dir flag shorthand = %q, want \"o\"", flag.Shorthand)
	}
	if flag.DefValue != "" {
		t.Errorf("output-dir flag default = %q, want \"\"", flag.DefValue)
	}
}

func TestRootCommand_VerboseFlag(t *testing.T) {
	cmd := rootCommand()
	flag := cmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Fatal("rootCommand missing --verbose persistent flag")
	}
	if flag.Shorthand != "v" {
		t.Errorf("verbose flag shorthand = %q, want \"v\"", flag.Shorthand)
	}
	if flag.DefValue != "false" {
		t.Errorf("verbose flag default = %q, want \"false\"", flag.DefValue)
	}
}
