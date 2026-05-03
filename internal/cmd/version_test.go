package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionVariables(t *testing.T) {
	orig := Version
	origTS := Timestamp
	t.Cleanup(func() {
		Version = orig
		Timestamp = origTS
	})

	Version = "1.2.3"
	Timestamp = "2026-05-02T00:00:00Z"

	if Version != "1.2.3" {
		t.Errorf("Version = %q, want '1.2.3'", Version)
	}
	if Timestamp != "2026-05-02T00:00:00Z" {
		t.Errorf("Timestamp = %q, want '2026-05-02T00:00:00Z'", Timestamp)
	}
}

func TestVersionInfo(t *testing.T) {
	vi := VersionInfo{Version: "1.0.0", Timestamp: "2026-05-02"}
	if vi.Version != "1.0.0" {
		t.Errorf("VersionInfo.Version = %q, want '1.0.0'", vi.Version)
	}
	if vi.Timestamp != "2026-05-02" {
		t.Errorf("VersionInfo.Timestamp = %q, want '2026-05-02'", vi.Timestamp)
	}
}

func TestVersionCommand_Structure(t *testing.T) {
	cmd := versionCommand()
	if cmd.Use != "version" {
		t.Errorf("Use = %q, want 'version'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short description must not be empty")
	}
}

// TestVersionSubcommand_Output runs a subprocess to test the version subcommand output
// since fmt.Printf writes directly to os.Stdout and cannot be redirected via cobra.
func TestVersionSubcommand_Output(t *testing.T) {
	if os.Getenv("TEST_VERSION_SUBPROCESS") == "1" {
		Version = "2.0.0"
		Timestamp = "2026-06-01"
		os.Args = []string{"backup-github", "version"} //nolint:reassign //Reassignment is for testing.
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestVersionSubcommand_Output", "-test.v")
	cmd.Env = append(os.Environ(), "TEST_VERSION_SUBPROCESS=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("subprocess failed: %v\nOutput:\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Version: 2.0.0") {
		t.Errorf("expected 'Version: 2.0.0' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Build Time: 2026-06-01") {
		t.Errorf("expected 'Build Time: 2026-06-01' in output, got:\n%s", output)
	}
}
