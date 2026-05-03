package cmd

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestLogoutCommand_Structure(t *testing.T) {
	cmd := logoutCommand()
	if cmd.Use != "logout" {
		t.Errorf("Use = %q, want 'logout'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short description must not be empty")
	}
}

func TestLogoutCommand_Execute_Success(t *testing.T) {
	keyring.MockInit()
	if err := keyring.Set("Backup GitHub Auth", "Ov23liy7HJJ5TtS55Eho", "ghp_testtoken"); err != nil {
		t.Fatalf("keyring.Set failed: %v", err)
	}

	cmd := rootCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"logout"})
	if err := cmd.Execute(); err != nil {
		t.Errorf("logout command error = %v, want nil", err)
	}
}

func TestLogoutCommand_Execute_NoToken(t *testing.T) {
	keyring.MockInit()

	cmd := rootCommand()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"logout"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when no token in keyring, got nil")
	}
}
