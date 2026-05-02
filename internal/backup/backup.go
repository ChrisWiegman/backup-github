package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ChrisWiegman/backup-github/internal/client"

	"github.com/google/go-github/v85/github"
)

func ExecuteBackup() error {
	repos, errCh := getRepos(client.GetGitHubClient())

	for repo := range repos {
		err := backupRepo(repo)
		if err != nil {
			return fmt.Errorf("error encountered in backing up repo %s: %w", repo.GetName(), err)
		}
	}

	err := <-errCh
	if err != nil {
		return fmt.Errorf("error encountered in retrieving all repos: %w", err)
	}

	return nil
}

func getRepos( //nolint:gocritic // Names aren't necessary in current context.
	ghClient *github.Client,
) (<-chan *github.Repository, <-chan error) {
	reposCh := make(chan *github.Repository)
	errCh := make(chan error, 1)

	go func() {
		defer close(reposCh)
		opts := &github.RepositoryListByAuthenticatedUserOptions{Type: "all"}

		for {
			repos, resp, err := ghClient.Repositories.ListByAuthenticatedUser(context.Background(), opts)
			if rateErr, ok := errors.AsType[*github.AbuseRateLimitError](err); ok {
				errCh <- fmt.Errorf("hit secondary rate limit, retry after %v", rateErr.RetryAfter)
				return
			}
			for _, repo := range repos {
				reposCh <- repo
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}
		errCh <- nil
	}()

	return reposCh, errCh
}

func backupRepo(repo *github.Repository) error {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not determine working directory: %w", err)
	}

	dest := filepath.Join(currentDirectory, "backups", filepath.Base(repo.GetName()))

	cmd := exec.CommandContext( //nolint:gosec // SSH URL is sourced from the authenticated GitHub API, not user input.
		context.Background(),
		"git",
		"clone",
		"--mirror",
		"--",
		repo.GetSSHURL(),
		dest,
	)

	_, err = os.Stat(dest)
	if !os.IsNotExist(err) {
		cmd = exec.CommandContext(context.Background(), "git", "-C", dest, "remote", "update")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
