package backup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ChrisWiegman/backup-github/internal/client"

	"github.com/google/go-github/v85/github"
)

func ExecuteBackup(verbose bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	verboseW := io.Discard
	if verbose {
		verboseW = os.Stdout
	}

	return executeBackup(ctx, os.Stdout, verboseW, client.GetGitHubClient(verboseW))
}

func executeBackup(ctx context.Context, w, verboseW io.Writer, ghClient *github.Client) error {
	repos, errCh := getRepos(ctx, ghClient, verboseW)

	var allRepos []*github.Repository
	for repo := range repos {
		allRepos = append(allRepos, repo)
	}

	if err := <-errCh; err != nil {
		return fmt.Errorf("error encountered in retrieving all repos: %w", err)
	}

	total := len(allRepos)
	fmt.Fprintf(verboseW, "Found %d repos to backup\n", total)

	const maxConcurrent = 5
	sem := make(chan struct{}, maxConcurrent)

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		counter atomic.Int64
		errs    []error
	)

	for _, repo := range allRepos {
		wg.Add(1)
		go func(repo *github.Repository) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			if err := backupRepo(ctx, w, verboseW, &mu, &counter, repo, total); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("error backing up %s: %w", repo.GetName(), err))
				mu.Unlock()
			}
		}(repo)
	}

	wg.Wait()
	return errors.Join(errs...)
}

func getRepos( //nolint:gocritic // Names aren't necessary in current context.
	ctx context.Context,
	ghClient *github.Client,
	verboseW io.Writer,
) (<-chan *github.Repository, <-chan error) {
	reposCh := make(chan *github.Repository)
	errCh := make(chan error, 1)

	go func() {
		defer close(reposCh)
		opts := &github.RepositoryListByAuthenticatedUserOptions{Type: "all"}
		pageNum := 1

		for {
			fmt.Fprintf(verboseW, "Fetching page %d of repositories...\n", pageNum)

			repos, resp, err := ghClient.Repositories.ListByAuthenticatedUser(ctx, opts)
			if err != nil {
				if rateErr, ok := errors.AsType[*github.AbuseRateLimitError](err); ok {
					errCh <- fmt.Errorf("hit secondary rate limit, retry after %v", rateErr.RetryAfter)
				} else {
					errCh <- err
				}
				return
			}
			for _, repo := range repos {
				reposCh <- repo
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
			pageNum++
		}
		errCh <- nil
	}()

	return reposCh, errCh
}

func backupRepo(
	ctx context.Context,
	w, verboseW io.Writer,
	mu *sync.Mutex,
	counter *atomic.Int64,
	repo *github.Repository,
	total int,
) error {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	dest := filepath.Join(currentDirectory, "backups", filepath.Base(repo.GetName()))

	cmd := exec.CommandContext( //nolint:gosec // SSH URL is sourced from the authenticated GitHub API, not user input.
		ctx,
		"git",
		"clone",
		"--mirror",
		"--",
		repo.GetSSHURL(),
		dest,
	)

	action := "Cloning"

	if _, err = os.Stat(dest); !os.IsNotExist(err) {
		action = "Updating"
		cmd = exec.CommandContext(ctx, "git", "-C", dest, "remote", "update")
	}

	n := counter.Add(1)
	mu.Lock()
	fmt.Fprintf(w, "[%d/%d] %s %s\n", n, total, action, repo.GetName())
	mu.Unlock()

	fmt.Fprintf(verboseW, "Running: %s\n", cmd.String())

	var stderr bytes.Buffer
	cmd.Stderr = io.MultiWriter(&stderr, verboseW)

	if err = cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%w\n%s", err, msg)
		}
		return err
	}
	return nil
}
