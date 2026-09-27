# gitea-cli

A command-line tool for operating a self-hosted Gitea server, written in Go. **Designed specifically for AI agents.**

> **🎉 This project was built entirely by AI**: including requirements analysis, architecture design, code development, unit testing, end-to-end testing (against a real Gitea server), and documentation — all done independently by AI (Claude), with no human-written code.

English (current) | [中文](README.zh-CN.md)

## Features

- Built on the official [Gitea Go SDK](https://code.gitea.io/sdk/gitea)
- **Credentials managed via git's own keyring (credential helper)** — never stores passwords itself
- Full CRUD for repositories, issues, PRs, users, organizations, teams, releases, and webhooks
- Configuration and credentials are separated: server info in a config file, token in the keyring
- **AI-friendly**: global `--json` flag for structured JSON output (including errors); automatically refuses destructive operations in non-interactive environments

## Installation

### Build from source

```bash
go build -o gitea-cli .
# Optional: install to PATH
go install
```

### Download from GitHub Release

This repository automatically builds cross-platform binaries via GitHub Actions. Each release generates executables for Linux / macOS / Windows and more, uploaded to the GitHub Release.

```bash
# Download the binary for your platform (example: v1.0.0, linux/amd64)
gh release download v1.0.0 --pattern 'gitea-cli_v1.0.0_linux_amd64.tar.gz'
tar xzf gitea-cli_v1.0.0_linux_amd64.tar.gz
sudo mv gitea-cli /usr/local/bin/
```

## Authentication

gitea-cli requires authentication to access the Gitea API. **The essence of authentication is simple: an access token is provided externally, and the tool sends it to the Gitea server via HTTP Basic Auth.**

What is called "initialization" is simply **one way of providing the token** — it stores the token in the git keyring so it can be read automatically on every subsequent invocation. Initialization, configuration, and authentication all serve the same single goal: letting gitea-cli pass the Gitea server's identity verification.

### Authentication mechanism

The tool authenticates via **HTTP Basic Auth**: the username is sent as the username, and the access token is sent as the password.

Basic Auth is used instead of the `Authorization: token` header for compatibility — some deployment environments (e.g. behind a WAF such as SafeLine) block the `Authorization: token` header but allow Basic Auth.

### Token sources (by priority)

The tool looks up the token in the following order and uses the first one found; if none is provided, authentication fails:

| Priority | Source | Description |
|----------|--------|-------------|
| 1 | Config file `token` field / `GITEA_TOKEN` env var | Explicitly set in config or environment |
| 2 | git keyring (credential helper) | Password retrieved via `git credential fill`, used as the token |

### Initialization (write the token to the keyring)

`config init` stores the token in the git keyring. This is the recommended approach — the token is not written to the config file, which is safer.

```bash
# Interactive: follow the prompts for server URL, username, and token
gitea-cli config init

# Non-interactive (recommended for AI agents): token from stdin, avoiding shell history
echo "$TOKEN" | gitea-cli config init --url https://gitea.example.com --username alice --token-stdin

# Or provide server URL and username via environment variables
export GITEA_URL=https://gitea.example.com
export GITEA_USERNAME=alice
printf '%s' "$TOKEN" | gitea-cli config init --token-stdin
```

The token is stored via `git credential approve` into the git keyring (e.g. macOS Keychain, Linux Secret Service, or `~/.git-credentials`), depending on your `git config credential.helper`. **The token is never written to the config file.**

### Keyring management

```bash
# Show current configuration and authentication status
gitea-cli config show

# Remove credentials from the keyring
gitea-cli config clear
```

## Configuration

Configuration is separated from authentication: **server URL and username live in the config file, while the token lives in the keyring** (optionally, it can also be placed in the config file, though this is not recommended).

### Config file

Stored at `~/.gitea-cli/config.yaml`:

```yaml
url: https://gitea.example.com
username: yourname
# token: (optional; usually omitted in favor of the keyring)
```

### Environment variables

Higher priority than the config file:

| Variable | Description |
|----------|-------------|
| `GITEA_URL` | Server URL |
| `GITEA_USERNAME` | Username |
| `GITEA_TOKEN` | Access token (optional) |

## Usage

### Repositories

```bash
gitea-cli repo list                    # List my repositories
gitea-cli repo list username           # List a user/org's repositories
gitea-cli repo info owner/repo         # Show details
gitea-cli repo create my-repo --description "desc" --private
gitea-cli repo delete owner/repo       # Requires confirmation; use --yes to skip
gitea-cli repo star owner/repo         # Star
gitea-cli repo unstar owner/repo       # Unstar
gitea-cli repo watch owner/repo        # Watch
gitea-cli repo unwatch owner/repo      # Unwatch

# Labels
gitea-cli repo label list owner/repo
gitea-cli repo label create owner/repo "bug" "#ff0000" --description "bug label"
```

### Issues

```bash
gitea-cli issue list owner/repo --state open
gitea-cli issue info owner/repo 1
gitea-cli issue create owner/repo "title" --body "body"
gitea-cli issue close owner/repo 1
gitea-cli issue reopen owner/repo 1
gitea-cli issue delete owner/repo 1    # Requires confirmation; use --yes to skip

# Comments
gitea-cli issue comment list owner/repo 1
gitea-cli issue comment add owner/repo 1 "comment text"
```

### Pull Requests

```bash
gitea-cli pr list owner/repo
gitea-cli pr info owner/repo 3
gitea-cli pr create owner/repo "title" --head feature-branch --base main
```

### Users

```bash
gitea-cli user info              # Show the current user
gitea-cli user info username     # Show a specific user
gitea-cli user search keyword
```

### Organizations and Teams

```bash
# Organizations
gitea-cli org list              # List my organizations
gitea-cli org list username     # List a user's organizations
gitea-cli org info orgname
gitea-cli org create neworg --description "desc"
gitea-cli org delete neworg      # Requires confirmation; use --yes to skip

# Organization members
gitea-cli org member list orgname
gitea-cli org member add orgname username --team Owners
gitea-cli org member remove orgname username

# Teams
gitea-cli org team list orgname
gitea-cli org team create orgname devteam --permission write
gitea-cli org team member list <team-id>
gitea-cli org team member add <team-id> username
gitea-cli org team member remove <team-id> username
```

### Releases

```bash
gitea-cli release list owner/repo
gitea-cli release create owner/repo v1.0.0 --title "v1.0.0" --note "release notes"
gitea-cli release delete owner/repo v1.0.0   # Requires confirmation; use --yes to skip
```

### Webhooks

```bash
gitea-cli webhook list owner/repo
gitea-cli webhook create owner/repo gitea https://example.com/hook --secret xxx
gitea-cli webhook delete owner/repo <hook-id>   # Requires confirmation; use --yes to skip
```

## Guide for AI agents

This tool is specially optimized for AI agents (such as Claude Code, pi, Cursor, etc.).

### Core principles

1. **Always pass the `--json` flag** — all commands output structured JSON for easy programmatic parsing.
2. **Destructive operations require `--yes`** — in a non-interactive environment (AI agent), without `--yes` the command errors out immediately rather than hanging waiting for input.
3. **Judge success by exit code** — `0` means success, non-zero means failure.

### Output format conventions

- **Success**: outputs JSON data (object or array), exit code `0`.
- **Failure**: outputs `{"ok": false, "error": "..."}` to stderr, exit code `1`.

### Non-interactive initialization

AI agents have no interactive terminal, so info must be provided via flags or environment variables (see [Authentication](#authentication) above):

```bash
echo "$TOKEN" | gitea-cli config init --url https://gitea.example.com --username alice --token-stdin
```

### Recommended usage examples

```bash
# List repositories (JSON)
gitea-cli repo list --json

# Show repository details
gitea-cli repo info owner/repo --json

# Create an issue
gitea-cli issue create owner/repo "title" --body "body" --json

# Non-interactive delete (--yes required)
gitea-cli repo delete owner/repo --yes
```

## Project structure

```
.
├── main.go                 # Entry point
├── cmd/                    # CLI command definitions
│   ├── root.go             # Root command
│   ├── config.go           # config subcommands
│   ├── repo.go             # Repository management (labels, stars, watches)
│   ├── issue.go            # Issue management (comments)
│   ├── pr.go               # PR management
│   ├── user.go             # User management
│   ├── org.go              # Organization management (members)
│   ├── org_team.go         # Organization team management
│   ├── release.go          # Release management
│   ├── webhook.go          # Webhook management
│   ├── helpers.go          # Helper functions
│   ├── confirm.go          # Deletion confirmation
│   └── password.go         # Terminal password reading
├── config/                 # Config loading/saving
├── client/                 # Gitea client wrapper
└── credential/             # git credential keyring interaction
```

## Releasing

This project automates releases via GitHub Actions + GoReleaser. Pushing a `v*` tag triggers the workflow, which automatically:

1. Runs unit tests
2. Cross-compiles multi-platform binaries (Linux / macOS / Windows / FreeBSD / OpenBSD × amd64 / arm64 / arm)
3. Creates a GitHub Release and uploads the built binaries, checksums, and source archives

### Releasing a new version (using gh)

```bash
# 1. Create and push a version tag (triggers the workflow)
git tag v1.0.0
git push origin v1.0.0

# 2. Wait for GitHub Actions to finish (about 1-2 minutes); check progress:
gh run watch

# 3. Once done, the Release is created automatically with binaries attached:
gh release view v1.0.0

# 4. (Optional) List the uploaded assets
gh release view v1.0.0 --json assets
```

> Note: the workflow uses GoReleaser; the version is taken from the git tag. Binaries embed the version, commit hash, and build date, viewable via `gitea-cli --version`.

## Dependencies

This project is built on the following excellent open-source projects:

| Dependency | Purpose | Link |
|------------|---------|------|
| Gitea Go SDK | Gitea API client | https://code.gitea.io/sdk/gitea |
| Cobra | CLI framework | https://github.com/spf13/cobra |
| Viper | Config management | https://github.com/spf13/viper |
| golang.org/x/term | Terminal detection | https://pkg.go.dev/golang.org/x/term |

## Acknowledgments

Thanks to the following projects and communities:

- **[Gitea](https://gitea.io/)** team — for the excellent self-hosted Git service and the officially maintained [Gitea Go SDK](https://code.gitea.io/sdk/gitea), on which this project is built.
- **[Cobra](https://github.com/spf13/cobra)** — a powerful Go CLI framework that makes building CLI tools clean and elegant.
- **[Viper](https://github.com/spf13/viper)** — a flexible Go configuration library.
- **git credential helper ecosystem** — this project reuses git's own keyring mechanism (libsecret / osxkeychain / pass / store, etc.) for credential management, following git's security conventions.

## License

MIT
