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
