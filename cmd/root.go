package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"gitea-cli/config"
)

var rootCmd = &cobra.Command{
	Use:   "gitea-cli",
	Short: "Gitea CLI - a command-line tool for operating a self-hosted Gitea server",
	Long: `gitea-cli is a command-line tool for operating a self-hosted Gitea server, designed for AI agents.

It supports full CRUD for repositories, issues, PRs, users, organizations, teams,
releases, and webhooks. Credentials are managed via git's own keyring
(credential helper), keeping secrets safe.

Global flags:
  --json    Output everything as JSON (including errors) for easy parsing by AI.
            Recommended for AI agents to always use.

Typical usage:
  gitea-cli repo list --json
  gitea-cli issue info owner/repo 1 --json
  gitea-cli repo create my-repo --json`,
	Version:       "dev",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// SetVersion sets the version string shown by --version.
// It must be called before Execute; it updates rootCmd.Version directly
// because cobra reads the Version field at command construction time.
func SetVersion(version, commit, date string) {
	if commit != "" && commit != "none" {
		rootCmd.Version = version + " (commit " + commit + ", built " + date + ")"
	} else {
		rootCmd.Version = version
	}
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadConfig loads the configuration, printing an error and exiting on failure.
func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		fail("%v", err)
	}
	return cfg
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "output JSON (including errors) for easy parsing by AI")
}
