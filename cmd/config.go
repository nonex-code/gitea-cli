package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gitea-cli/config"
	"gitea-cli/credential"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Gitea server info and credentials",
	Long:  "Configure the Gitea server URL and username, and store credentials via the git keyring.",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config and store credentials in the git keyring",
	Long: `Initialize configuration and store credentials.

Two modes are supported:
1. Interactive: run gitea-cli config init and follow the prompts.
2. Non-interactive (recommended for AI agents): provide info via flags or env vars.

Non-interactive examples:
  # via flags (token read from stdin, avoiding shell history)
  echo "$TOKEN" | gitea-cli config init --url https://gitea.example.com --username alice

  # via environment variables
  GITEA_URL=https://gitea.example.com GITEA_USERNAME=alice \
    gitea-cli config init --token-stdin < token.txt

Note: the token is only stored in the git keyring (credential helper),
never written to the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fail("%v", err)
		}

		// Command-line flags override (non-interactive mode)
		if v, _ := cmd.Flags().GetString("url"); v != "" {
			cfg.URL = v
		}
		if v, _ := cmd.Flags().GetString("username"); v != "" {
			cfg.Username = v
		}

		// If still missing, try reading from stdin (one field per line: url, username, token)
		tokenFromFlag, _ := cmd.Flags().GetString("token")
		tokenStdin, _ := cmd.Flags().GetBool("token-stdin")

		if cfg.URL == "" || cfg.Username == "" {
			if !termIsTTY() {
				// Non-interactive and incomplete args: try reading line by line from stdin
				scanner := bufio.NewScanner(os.Stdin)
				if cfg.URL == "" && scanner.Scan() {
					cfg.URL = strings.TrimSpace(scanner.Text())
				}
				if cfg.Username == "" && scanner.Scan() {
					cfg.Username = strings.TrimSpace(scanner.Text())
				}
			} else {
				if cfg.URL == "" {
					fmt.Print("Gitea server URL (e.g. https://gitea.example.com): ")
					fmt.Scanln(&cfg.URL)
				}
				if cfg.Username == "" {
					fmt.Print("Username: ")
					fmt.Scanln(&cfg.Username)
				}
			}
		}

		if cfg.URL == "" || cfg.Username == "" {
			fail("server URL and username are required; provide them via --url/--username or environment variables")
		}

		// Read token: prefer --token flag, then --token-stdin (from stdin), finally interactive
		var token string
		switch {
		case tokenFromFlag != "":
			token = tokenFromFlag
		case tokenStdin:
			b, err := readAllStdin()
			if err != nil {
				fail("failed to read token from stdin: %v", err)
			}
			token = strings.TrimSpace(b)
		default:
			if !termIsTTY() {
				fail("non-interactive environment requires an explicit token; use --token or --token-stdin")
			}
			fmt.Print("Access token (input hidden): ")
			token, err = readPassword()
			if err != nil {
				fail("failed to read token: %v", err)
			}
			fmt.Println()
		}

		if token == "" {
			fail("token cannot be empty")
		}

		// Verify the token FIRST by making a real API call. Only store it to the
		// keyring after it is confirmed valid, so a wrong token never overwrites
		// good credentials.
		if err := verifyToken(cfg.URL, cfg.Username, token); err != nil {
			fail("token verification failed: %v", err)
		}

		// Save config (token is NOT saved; it goes to the keyring)
		if err := config.Save(&config.Config{URL: cfg.URL, Username: cfg.Username}); err != nil {
			fail("%v", err)
		}

		// Store the verified credentials into the git keyring
		if err := credential.Store(cfg.URL, cfg.Username, token); err != nil {
			fail("%v", err)
		}

		fmt.Printf("✓ Config saved to ~/.gitea-cli/config.yaml\n")
		fmt.Printf("✓ Credentials stored and verified (%s)\n", cfg.URL)
	},
}

// verifyToken makes a real HTTP request to the Gitea API to confirm the token
// is valid. It distinguishes two cases:
//   - HTTP 401: authentication failed — the username/token is wrong.
//   - HTTP 403: authenticated but insufficient scope — the token is valid, just
//     limited (which is an intentional minimal-permission setup).
//
// Only a 401 (or a network/transport error) is treated as a failure. A 403 is
// considered success because the token itself is correct.
func verifyToken(serverURL, username, token string) error {
	u := strings.TrimRight(serverURL, "/") + "/api/v1/user"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid username, password or token")
	}
	// 403 (insufficient scope) and 2xx (full access) both mean the token is valid.
	return nil
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fail("%v", err)
		}
		if jsonEnabled() {
			printJSON(map[string]interface{}{
				"url":       cfg.URL,
				"username":  cfg.Username,
				"has_token": cfg.Token != "",
				"token_in":  tokenLocation(cfg),
			})
			return
		}
		fmt.Printf("Server URL: %s\n", cfg.URL)
		fmt.Printf("Username:   %s\n", cfg.Username)
		if cfg.Token != "" {
			fmt.Println("Token:      configured (from config file)")
		} else {
			fmt.Println("Token:      not configured (will use git keyring)")
		}
	},
}

// tokenLocation returns a description of where the token comes from.
func tokenLocation(cfg *config.Config) string {
	if cfg.Token != "" {
		return "config"
	}
	return "credential-helper"
}

var configClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove credentials from the keyring",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fail("%v", err)
		}
		if cfg.URL == "" {
			fail("server URL is not configured")
		}
		if err := credential.Erase(cfg.URL, cfg.Username); err != nil {
			fail("%v", err)
		}
		fmt.Printf("✓ Credentials for %s removed from the keyring\n", cfg.URL)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configClearCmd)

	configInitCmd.Flags().String("url", "", "Gitea server URL")
	configInitCmd.Flags().String("username", "", "Username")
	configInitCmd.Flags().String("token", "", "Access token (not recommended; appears in shell history)")
	configInitCmd.Flags().Bool("token-stdin", false, "Read token from stdin (recommended)")
}
