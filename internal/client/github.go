package client

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cli/oauth"
	"github.com/google/go-github/v85/github"
	"github.com/zalando/go-keyring"
)

const clientID = "Ov23liy7HJJ5TtS55Eho"
const serviceName = "Backup GitHub Auth"

func GetGitHubClient(verboseW io.Writer) *github.Client {
	token, _ := getGitHubAuth(verboseW)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return github.NewClient(httpClient).WithAuthToken(token)
}

func LogoutGitHub() error {
	return keyring.Delete(serviceName, clientID)
}

func getGitHubAuth(verboseW io.Writer) (string, error) {
	token, err := keyring.Get(serviceName, clientID)
	if err != nil {
		fmt.Fprintln(verboseW, "No stored token found, starting OAuth flow...")

		host, hostErr := oauth.NewGitHubHost("https://github.com")
		if hostErr != nil {
			return "", hostErr
		}

		flow := &oauth.Flow{
			Host:     host,
			ClientID: clientID,
			Scopes:   []string{"repo", "read:org", "gist"},
		}

		accessToken, flowErr := flow.DetectFlow()
		if flowErr != nil {
			return "", flowErr
		}

		fmt.Fprintln(verboseW, "OAuth flow completed, storing token.")
		return accessToken.Token, keyring.Set(serviceName, clientID, accessToken.Token)
	}

	fmt.Fprintln(verboseW, "Found stored token in keyring.")
	return token, nil
}
