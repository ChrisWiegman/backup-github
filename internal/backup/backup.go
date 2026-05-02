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

func ExecuteBackup() {
	repos, err := getRepos()
	if err != nil {
		panic(err)
	}

	for _, repo := range repos {
		_ = cloneRepo(repo)
	}
}

func getRepos() ([]*github.Repository, error) {
	ghClient := client.GetGitHubClient()

	var allRepos []*github.Repository

	opts := &github.RepositoryListByAuthenticatedUserOptions{
		Type: "all",
	}

	for {
		repos, resp, err := ghClient.Repositories.ListByAuthenticatedUser(context.Background(), opts)
		var rateErr *github.AbuseRateLimitError
		if errors.As(err, &rateErr) {
			return allRepos, fmt.Errorf("hit secondary rate limit, retry after %v", rateErr.RetryAfter)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func cloneRepo(repo *github.Repository) error {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
