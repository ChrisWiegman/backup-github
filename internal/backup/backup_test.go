package backup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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
		out, err := exec.Command(
			args[0],
			args[1:]...,
		).
			CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	destDir := t.TempDir()
	mirrorPath := filepath.Join(destDir, "test-repo")

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

	if err = backupRepo(
		context.Background(),
		io.Discard,
		io.Discard,
		&sync.Mutex{},
		&atomic.Int64{},
		repo,
		1,
		"",
		false,
	); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

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
		out, err := exec.CommandContext(context.Background(), args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	destDir := t.TempDir()
	mirrorPath := filepath.Join(destDir, "test-repo")

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
	if err = backupRepo(
		context.Background(),
		io.Discard,
		io.Discard,
		&sync.Mutex{},
		&atomic.Int64{},
		repo,
		1,
		"",
		false,
	); err != nil {
		t.Fatalf("initial backupRepo returned error: %v", err)
	}

	// Second call: destination already exists, should run remote update instead.
	if err = backupRepo(
		context.Background(),
		io.Discard,
		io.Discard,
		&sync.Mutex{},
		&atomic.Int64{},
		repo,
		1,
		"",
		false,
	); err != nil {
		t.Fatalf("update backupRepo returned error: %v", err)
	}

	if _, err = os.Stat(mirrorPath); os.IsNotExist(err) {
		t.Errorf("expected mirror clone at %s, but it was not found after update", mirrorPath)
	}
}

func TestCloneRepo_InvalidURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	repo := &github.Repository{
		Name:   new("bad-repo"),
		SSHURL: new("/this/path/does/not/exist"),
	}

	if err := backupRepo(
		context.Background(),
		io.Discard,
		io.Discard,
		&sync.Mutex{},
		&atomic.Int64{},
		repo,
		1,
		"",
		false,
	); err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
}

func TestBackupRepo_VerboseOutput(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
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

	var verboseBuf bytes.Buffer
	if err = backupRepo(
		context.Background(),
		io.Discard,
		&verboseBuf,
		&sync.Mutex{},
		&atomic.Int64{},
		repo,
		1,
		"",
		false,
	); err != nil {
		t.Fatalf("backupRepo returned error: %v", err)
	}

	if !strings.Contains(verboseBuf.String(), "Running: ") {
		t.Errorf("expected git command in verbose output, got: %s", verboseBuf.String())
	}
}

func newTestGitHubClient(t *testing.T, mux *http.ServeMux) *github.Client {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	ghClient := github.NewClient(nil)
	ghClient.BaseURL, _ = url.Parse(server.URL + "/")
	return ghClient
}

func TestGetRepos_SinglePage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) { //nolint:revive //Issue in tests.
		fmt.Fprint(
			w,
			`[{"name":"repo1","ssh_url":"git@github.com:user/repo1.git"},{"name":"repo2","ssh_url":"git@github.com:user/repo2.git"}]`,
		)
	})

	reposCh, errCh := getRepos(context.Background(), newTestGitHubClient(t, mux), io.Discard)

	var names []string
	for repo := range reposCh {
		names = append(names, repo.GetName())
	}

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 repos, got %d: %v", len(names), names)
	}
}

func TestGetRepos_Paginated(t *testing.T) {
	mux := http.NewServeMux()
	var serverURL string
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `[{"name":"repo2"}]`)
		} else {
			w.Header().Set("Link", fmt.Sprintf(`<%s/user/repos?page=2>; rel="next"`, serverURL))
			fmt.Fprint(w, `[{"name":"repo1"}]`)
		}
	})

	ghClient := newTestGitHubClient(t, mux)
	serverURL = ghClient.BaseURL.String()

	reposCh, errCh := getRepos(context.Background(), ghClient, io.Discard)

	var names []string
	for repo := range reposCh {
		names = append(names, repo.GetName())
	}

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 repos, got %d: %v", len(names), names)
	}
}

func TestGetRepos_RateLimit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) { //nolint:revive //Issue in tests.
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(
			w,
			`{"message":"You have exceeded a secondary rate limit.","documentation_url":"https://docs.github.com/rest/overview/resources-in-the-rest-api#secondary-rate-limits"}`,
		)
	})

	_, errCh := getRepos(context.Background(), newTestGitHubClient(t, mux), io.Discard)

	if err := <-errCh; err == nil {
		t.Error("expected rate limit error, got nil")
	}
}

func TestGetRepos_VerboseLogging(t *testing.T) {
	mux := http.NewServeMux()
	var serverURL string
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `[{"name":"repo2"}]`)
		} else {
			w.Header().Set("Link", fmt.Sprintf(`<%s/user/repos?page=2>; rel="next"`, serverURL))
			fmt.Fprint(w, `[{"name":"repo1"}]`)
		}
	})

	ghClient := newTestGitHubClient(t, mux)
	serverURL = ghClient.BaseURL.String()

	var verboseBuf bytes.Buffer
	reposCh, errCh := getRepos(context.Background(), ghClient, &verboseBuf)

	for range reposCh { //nolint:revive //This is needed to prevent hanging tests.
	}

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := verboseBuf.String()
	if !strings.Contains(output, "page 1") {
		t.Errorf("expected page 1 in verbose output, got: %s", output)
	}
	if !strings.Contains(output, "page 2") {
		t.Errorf("expected page 2 in verbose output, got: %s", output)
	}
}

func TestExecuteBackup_Progress(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) { //nolint:revive //Issue in tests.
		fmt.Fprintf(w, `[{"name":"repo1","ssh_url":%q}]`, srcDir)
	})
	ghClient := newTestGitHubClient(t, mux)

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

	var buf bytes.Buffer
	if err = executeBackup(context.Background(), &buf, io.Discard, ghClient, "", false); err != nil {
		t.Fatalf("executeBackup returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "[1/1] Cloning repo1") {
		t.Errorf("expected Cloning progress line, got: %s", buf.String())
	}

	buf.Reset()
	if err = executeBackup(context.Background(), &buf, io.Discard, ghClient, "", false); err != nil {
		t.Fatalf("second executeBackup returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "[1/1] Updating repo1") {
		t.Errorf("expected Updating progress line, got: %s", buf.String())
	}
}

func TestExecuteBackup_VerboseOutput(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) { //nolint:revive //Issue in tests.
		fmt.Fprintf(w, `[{"name":"repo1","ssh_url":%q}]`, srcDir)
	})
	ghClient := newTestGitHubClient(t, mux)

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

	var verboseBuf bytes.Buffer
	if err = executeBackup(context.Background(), io.Discard, &verboseBuf, ghClient, "", false); err != nil {
		t.Fatalf("executeBackup returned error: %v", err)
	}

	output := verboseBuf.String()
	if !strings.Contains(output, "Found 1 repos to backup") {
		t.Errorf("expected repo count in verbose output, got: %s", output)
	}
	if !strings.Contains(output, "Running: ") {
		t.Errorf("expected git command in verbose output, got: %s", output)
	}
}

func TestBackupRepo_OutputDir_NoCwd(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	repo := &github.Repository{Name: new("test-repo"), SSHURL: new(srcDir)}
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

	if err = backupRepo(
		context.Background(), io.Discard, io.Discard, &sync.Mutex{}, &atomic.Int64{}, repo, 1, "", false,
	); err != nil {
		t.Fatalf("backupRepo returned error: %v", err)
	}
	if _, err = os.Stat(filepath.Join(destDir, "test-repo")); os.IsNotExist(err) {
		t.Errorf("expected clone in cwd, not found")
	}
}

func TestBackupRepo_OutputDir_Absolute(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	repo := &github.Repository{Name: new("test-repo"), SSHURL: new(srcDir)}
	destDir := t.TempDir()

	if err := backupRepo(
		context.Background(), io.Discard, io.Discard, &sync.Mutex{}, &atomic.Int64{}, repo, 1, destDir, true,
	); err != nil {
		t.Fatalf("backupRepo returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "test-repo")); os.IsNotExist(err) {
		t.Errorf("expected clone at absolute path, not found")
	}
}

func TestBackupRepo_OutputDir_Relative(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	repo := &github.Repository{Name: new("test-repo"), SSHURL: new(srcDir)}
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

	if err = backupRepo(
		context.Background(), io.Discard, io.Discard, &sync.Mutex{}, &atomic.Int64{}, repo, 1, "backups", true,
	); err != nil {
		t.Fatalf("backupRepo returned error: %v", err)
	}
	if _, err = os.Stat(filepath.Join(destDir, "backups", "test-repo")); os.IsNotExist(err) {
		t.Errorf("expected clone at relative path, not found")
	}
}

func TestBackupRepo_OutputDir_HomeRelative(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not determine home directory")
	}

	repo := &github.Repository{Name: new("test-repo"), SSHURL: new(srcDir)}
	subDir := filepath.Join(homeDir, "backup-github-test-HomeRelative")
	t.Cleanup(func() { os.RemoveAll(subDir) })

	if err = backupRepo(
		context.Background(), io.Discard, io.Discard, &sync.Mutex{}, &atomic.Int64{}, repo, 1,
		"~/backup-github-test-HomeRelative", true,
	); err != nil {
		t.Fatalf("backupRepo returned error: %v", err)
	}
	if _, err = os.Stat(filepath.Join(subDir, "test-repo")); os.IsNotExist(err) {
		t.Errorf("expected clone at home-relative path, not found")
	}
}

func TestGetOutputDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("could not determine home directory")
	}

	const currentDir = "/some/current/dir"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "home-relative with slash",
			input: "~/backups",
			want:  filepath.Join(homeDir, "backups"),
		},
		{
			name:  "home-relative without slash",
			input: "~backups",
			want:  filepath.Join(homeDir, "backups"),
		},
		{
			name:  "absolute path",
			input: "/absolute/path",
			want:  "/absolute/path",
		},
		{
			name:  "relative path",
			input: "backups",
			want:  filepath.Join(currentDir, "backups"),
		},
		{
			name:  "dot-relative path",
			input: "./backups",
			want:  filepath.Join(currentDir, "backups"),
		},
		{
			name:  "parent-relative path",
			input: "../backups",
			want:  "/some/current/backups",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			got, err = getOutputDir(currentDir, tt.input)
			if err != nil {
				t.Fatalf("getOutputDir returned error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
