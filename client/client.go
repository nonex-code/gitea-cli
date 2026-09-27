package client

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"gitea-cli/config"
	"gitea-cli/credential"
)

// Client wraps the Gitea SDK client and additionally retains credential
// information for future use on APIs not yet covered by the SDK.
type Client struct {
	*gitea.Client
	baseURL  string
	username string
	token    string
}

// New creates a Gitea client.
// Credential resolution order:
//  1. Token explicitly configured in the config file
//  2. Password obtained from git credential (keyring) used as the token
//
// Basic auth (username + token) is used by default to remain compatible
// with deployments behind a WAF (some WAFs block the `Authorization: token`
// header but allow Basic auth).
func New(cfg *config.Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("Gitea server URL is not configured; run `gitea-cli config` or set the GITEA_URL environment variable first")
	}

	token := cfg.Token
	username := cfg.Username

	if token == "" {
		cred, err := credential.Get(cfg.URL, username)
		if err != nil {
			return nil, fmt.Errorf("failed to get credentials: %w", err)
		}
		token = cred.Password
		if username == "" {
			username = cred.Username
		}
	}

	// Use Basic auth (username + token as password) for maximum
	// compatibility across deployment environments (e.g. behind a WAF).
	opts := []gitea.ClientOption{
		gitea.SetBasicAuth(username, token),
		// Skip version probing during initialization (/api/v1/version).
		// That request carries no auth header and gets blocked by WAFs;
		// version probing is also not required for normal usage.
		gitea.SetGiteaVersion(""),
	}

	cli, err := gitea.NewClient(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gitea client: %w", err)
	}

	return &Client{
		Client:   cli,
		baseURL:  cfg.URL,
		username: username,
		token:    token,
	}, nil
}
