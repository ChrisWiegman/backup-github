package client

import (
	"github.com/cli/oauth"
	"github.com/google/go-github/v85/github"
	"github.com/zalando/go-keyring"
)

const clientID = "Ov23liy7HJJ5TtS55Eho"
const serviceName = "Backup GitHub Auth"

func GetGitHubClient() *github.Client {
	token, _ := getGitHubAuth()
	return github.NewClient(nil).WithAuthToken(token)
}

func getGitHubAuth() (string, error) {
	token, err := keyring.Get(serviceName, clientID)
	if err != nil {
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

		return accessToken.Token, keyring.Set(serviceName, clientID, accessToken.Token)
	}

	return token, nil
}
