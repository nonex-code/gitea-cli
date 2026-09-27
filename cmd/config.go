package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gitea-cli/config"
	"gitea-cli/credential"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置 Gitea 服务器信息与凭据",
	Long:  "配置 Gitea 服务器地址、用户名，并通过 git 密钥箱存储凭据。",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置并存储凭据到 git 密钥箱",
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
					fmt.Print("Gitea 服务器地址 (如 https://gitea.example.com): ")
					fmt.Scanln(&cfg.URL)
				}
				if cfg.Username == "" {
					fmt.Print("用户名: ")
					fmt.Scanln(&cfg.Username)
				}
			}
		}

		if cfg.URL == "" || cfg.Username == "" {
			fail("服务器地址和用户名不能为空，请通过 --url/--username 参数或环境变量提供")
		}

		// Read token: prefer --token flag, then --token-stdin (from stdin), finally interactive
		var token string
		switch {
		case tokenFromFlag != "":
			token = tokenFromFlag
		case tokenStdin:
			b, err := readAllStdin()
			if err != nil {
				fail("从 stdin 读取 token 失败: %v", err)
			}
			token = strings.TrimSpace(b)
		default:
			if !termIsTTY() {
				fail("非交互环境需要显式提供 token，请使用 --token 或 --token-stdin 参数")
			}
			fmt.Print("访问令牌 (Token，输入不可见): ")
			token, err = readPassword()
			if err != nil {
				fail("读取令牌失败: %v", err)
			}
			fmt.Println()
		}

		if token == "" {
			fail("令牌不能为空")
		}

		// Save config (token is NOT saved; it goes to the keyring)
		if err := config.Save(&config.Config{URL: cfg.URL, Username: cfg.Username}); err != nil {
			fail("%v", err)
		}

		// Store credentials into the git keyring
		if err := credential.Store(cfg.URL, cfg.Username, token); err != nil {
			fail("%v", err)
		}

		fmt.Printf("✓ 配置已保存到 ~/.gitea-cli/config.yaml\n")
		fmt.Printf("✓ 凭据已存储到 git 密钥箱 (%s)\n", cfg.URL)
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
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
		fmt.Printf("服务器地址: %s\n", cfg.URL)
		fmt.Printf("用户名:     %s\n", cfg.Username)
		if cfg.Token != "" {
			fmt.Println("Token:      已配置 (来自配置文件)")
		} else {
			fmt.Println("Token:      未配置 (将从 git 密钥箱获取)")
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
	Short: "清除密钥箱中的凭据",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fail("%v", err)
		}
		if cfg.URL == "" {
			fail("未配置服务器地址")
		}
		if err := credential.Erase(cfg.URL, cfg.Username); err != nil {
			fail("%v", err)
		}
		fmt.Printf("✓ 已从密钥箱清除 %s 的凭据\n", cfg.URL)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configClearCmd)

	configInitCmd.Flags().String("url", "", "Gitea 服务器地址")
	configInitCmd.Flags().String("username", "", "用户名")
	configInitCmd.Flags().String("token", "", "访问令牌（不推荐，会出现在 shell 历史中）")
	configInitCmd.Flags().Bool("token-stdin", false, "从 stdin 读取 token（推荐）")
}
