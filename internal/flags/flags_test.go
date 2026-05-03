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
