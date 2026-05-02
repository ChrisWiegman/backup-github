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
		host, err := oauth.NewGitHubHost("https://github.com")
		if err != nil {
			return "", err
		}

		flow := &oauth.Flow{
			Host:     host,
			ClientID: clientID,
			Scopes:   []string{"repo", "read:org", "gist"},
		}

		accessToken, err := flow.DetectFlow()
		if err != nil {
			return "", err
		}

		return accessToken.Token, keyring.Set(serviceName, clientID, accessToken.Token)
	}

	return token, nil
}
