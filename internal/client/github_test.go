package client

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestGetGitHubAuth_RetrievesTokenFromKeyring(t *testing.T) {
	keyring.MockInit()

	const wantToken = "ghp_testtoken123"
	if err := keyring.Set(serviceName, clientID, wantToken); err != nil {
		t.Fatalf("keyring.Set failed: %v", err)
	}

	token, err := getGitHubAuth()
	if err != nil {
		t.Fatalf("getGitHubAuth() error = %v, want nil", err)
	}
	if token != wantToken {
		t.Errorf("getGitHubAuth() = %q, want %q", token, wantToken)
	}
}

func TestGetGitHubClient_ReturnsNonNilClient(t *testing.T) {
	keyring.MockInit()

	if err := keyring.Set(serviceName, clientID, "ghp_testtoken"); err != nil {
		t.Fatalf("keyring.Set failed: %v", err)
	}

	client := GetGitHubClient()
	if client == nil {
		t.Error("GetGitHubClient() returned nil, want non-nil *github.Client")
	}
}

func TestGetGitHubAuth_KeyringConstants(t *testing.T) {
	if clientID == "" {
		t.Error("clientID must not be empty")
	}
	if serviceName == "" {
		t.Error("serviceName must not be empty")
	}
}
