package backup

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v85/github"
)

func TestCloneRepo_Success(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command( ///nolint:gosec // SSH URL is sourced from the authenticated GitHub API, not user input.
			args[0],
			args[1:]...,
		).
			CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	destDir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(destDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err = os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	repo := &github.Repository{
		Name:   new("test-repo"),
		SSHURL: new(srcDir),
	}

	if err = backupRepo(repo); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

	mirrorPath := filepath.Join(destDir, "backups", "test-repo")
	if _, err = os.Stat(mirrorPath); os.IsNotExist(err) {
		t.Errorf("expected mirror clone at %s, but it was not created", mirrorPath)
	}
}

func TestUpdateRepo_Success(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput() //nolint:gosec
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	destDir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(destDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err = os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	repo := &github.Repository{
		Name:   new("test-repo"),
		SSHURL: new(srcDir),
	}

	// First call: clone the repo.
	if err = backupRepo(repo); err != nil {
		t.Fatalf("initial backupRepo returned error: %v", err)
	}

	// Second call: destination already exists, should run remote update instead.
	if err = backupRepo(repo); err != nil {
		t.Fatalf("update backupRepo returned error: %v", err)
	}

	mirrorPath := filepath.Join(destDir, "backups", "test-repo")
	if _, err = os.Stat(mirrorPath); os.IsNotExist(err) {
		t.Errorf("expected mirror clone at %s, but it was not found after update", mirrorPath)
	}
}

func TestCloneRepo_InvalidURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	destDir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(destDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err = os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	repo := &github.Repository{
		Name:   new("bad-repo"),
		SSHURL: new("/this/path/does/not/exist"),
	}

	if err = backupRepo(repo); err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
}
