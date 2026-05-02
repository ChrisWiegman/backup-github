package backup

import (
	"encoding/json"
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

func newTestClient(t *testing.T, handler http.Handler) (*github.Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := github.NewClient(nil)
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = baseURL
	client.UploadURL = baseURL
	return client, server
}

func TestGetRepos_Empty(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "[]")
	}))

	repos, err := getRepos(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestGetRepos_MultipleRepos(t *testing.T) {
	want := []map[string]string{
		{"name": "repo1", "ssh_url": "git@github.com:user/repo1.git"},
		{"name": "repo2", "ssh_url": "git@github.com:user/repo2.git"},
	}

	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(want); err != nil {
			t.Error(err)
		}
	}))

	repos, err := getRepos(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].GetName() != "repo1" {
		t.Errorf("repos[0].Name = %q, want 'repo1'", repos[0].GetName())
	}
	if repos[1].GetName() != "repo2" {
		t.Errorf("repos[1].Name = %q, want 'repo2'", repos[1].GetName())
	}
}

func TestGetRepos_Pagination(t *testing.T) {
	page1 := []map[string]string{{"name": "repo1"}}
	page2 := []map[string]string{{"name": "repo2"}, {"name": "repo3"}}

	var serverURL string
	callCount := 0

	client, server := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			if err := json.NewEncoder(w).Encode(page2); err != nil {
				t.Error(err)
			}
			return
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s/user/repos?page=2>; rel="next"`, serverURL))
		if err := json.NewEncoder(w).Encode(page1); err != nil {
			t.Error(err)
		}
	}))
	serverURL = server.URL

	repos, err := getRepos(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 3 {
		t.Errorf("expected 3 repos across 2 pages, got %d", len(repos))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestGetRepos_AbuseRateLimit(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusForbidden)
		body := `{"message":"You have exceeded a secondary rate limit",` +
			`"documentation_url":"https://docs.github.com/rest/overview/` +
			`rate-limits-for-the-rest-api#secondary-rate-limits"}`
		fmt.Fprint(w, body)
	}))

	_, err := getRepos(client)
	if err == nil {
		t.Fatal("expected error for rate limit response, got nil")
	}
	expected := "hit secondary rate limit"
	if len(err.Error()) < len(expected) || err.Error()[:len(expected)] != expected {
		t.Errorf("error = %q, want message starting with %q", err.Error(), expected)
	}
}

func TestCloneRepo_Success(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available in PATH")
	}

	srcDir := t.TempDir()
	for _, args := range [][]string{
		{"git", "init", srcDir},
		{"git", "-C", srcDir, "-c", "user.email=test@test.com", "-c", "user.name=Test", "commit", "--allow-empty", "-m", "init"},
	} {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil { //nolint:gosec
			t.Skipf("git setup failed: %v\n%s", err, out)
		}
	}

	destDir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(destDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	repo := &github.Repository{
		Name:   github.Ptr("test-repo"),
		SSHURL: github.Ptr(srcDir),
	}

	if err := cloneRepo(repo); err != nil {
		t.Fatalf("cloneRepo returned error: %v", err)
	}

	mirrorPath := filepath.Join(destDir, "backups", "test-repo")
	if _, err := os.Stat(mirrorPath); os.IsNotExist(err) {
		t.Errorf("expected mirror clone at %s, but it was not created", mirrorPath)
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
	if err := os.Chdir(destDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	repo := &github.Repository{
		Name:   github.Ptr("bad-repo"),
		SSHURL: github.Ptr("/this/path/does/not/exist"),
	}

	if err := cloneRepo(repo); err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
}
