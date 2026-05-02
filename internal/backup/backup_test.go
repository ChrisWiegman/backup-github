package backup

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		out, err := exec.CommandContext(context.Background(), args[0], args[1:]...).CombinedOutput()
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

	reposCh, errCh := getRepos(newTestGitHubClient(t, mux))

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

	reposCh, errCh := getRepos(ghClient)

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

	_, errCh := getRepos(newTestGitHubClient(t, mux))

	if err := <-errCh; err == nil {
		t.Error("expected rate limit error, got nil")
	}
}
