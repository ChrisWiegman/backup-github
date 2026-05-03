package flags

import (
	"testing"

	"github.com/spf13/cobra"
)

func newTestCommand() *cobra.Command {
	return &cobra.Command{Use: "test"}
}

func TestAddVerboseFlag_RegistersFlag(t *testing.T) {
	cmd := newTestCommand()
	AddVerboseFlag(cmd)

	f := cmd.PersistentFlags().Lookup("verbose")
	if f == nil {
		t.Fatal("expected --verbose flag to be registered")
	}
	if f.Shorthand != "v" {
		t.Errorf("expected shorthand 'v', got %q", f.Shorthand)
	}
	if f.DefValue != "false" {
		t.Errorf("expected default value 'false', got %q", f.DefValue)
	}
}

func TestAddVerboseFlag_LongFlag(t *testing.T) {
	Verbose = false
	t.Cleanup(func() { Verbose = false })

	cmd := newTestCommand()
	AddVerboseFlag(cmd)

	if err := cmd.ParseFlags([]string{"--verbose"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !Verbose {
		t.Error("expected Verbose to be true after --verbose")
	}
}

func TestAddVerboseFlag_ShortFlag(t *testing.T) {
	Verbose = false
	t.Cleanup(func() { Verbose = false })

	cmd := newTestCommand()
	AddVerboseFlag(cmd)

	if err := cmd.ParseFlags([]string{"-v"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !Verbose {
		t.Error("expected Verbose to be true after -v")
	}
}

func TestAddOutputFlag_RegistersFlag(t *testing.T) {
	cmd := newTestCommand()
	AddOutputFlag(cmd)

	f := cmd.Flags().Lookup("output-dir")
	if f == nil {
		t.Fatal("expected --output-dir flag to be registered")
	}
	if f.Shorthand != "o" {
		t.Errorf("expected shorthand 'o', got %q", f.Shorthand)
	}
	if f.DefValue != "" {
		t.Errorf("expected empty default value, got %q", f.DefValue)
	}
}

func TestAddOutputFlag_LongFlag(t *testing.T) {
	OutputDir = ""
	t.Cleanup(func() { OutputDir = "" })

	cmd := newTestCommand()
	AddOutputFlag(cmd)

	if err := cmd.ParseFlags([]string{"--output-dir", "/tmp/backups"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if OutputDir != "/tmp/backups" {
		t.Errorf("expected OutputDir to be '/tmp/backups', got %q", OutputDir)
	}
}

func TestAddOutputFlag_ShortFlag(t *testing.T) {
	OutputDir = ""
	t.Cleanup(func() { OutputDir = "" })

	cmd := newTestCommand()
	AddOutputFlag(cmd)

	if err := cmd.ParseFlags([]string{"-o", "/tmp/backups"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if OutputDir != "/tmp/backups" {
		t.Errorf("expected OutputDir to be '/tmp/backups', got %q", OutputDir)
	}
}

func TestAddOutputFlag_DefaultEmpty(t *testing.T) {
	OutputDir = ""
	t.Cleanup(func() { OutputDir = "" })

	cmd := newTestCommand()
	AddOutputFlag(cmd)

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if OutputDir != "" {
		t.Errorf("expected OutputDir to remain empty when flag is not passed, got %q", OutputDir)
	}
}

func TestAddVerboseFlag_DefaultFalse(t *testing.T) {
	Verbose = false
	t.Cleanup(func() { Verbose = false })

	cmd := newTestCommand()
	AddVerboseFlag(cmd)

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if Verbose {
		t.Error("expected Verbose to remain false when flag is not passed")
	}
}
