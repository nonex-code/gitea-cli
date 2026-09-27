package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"gitea-cli/config"
)

var rootCmd = &cobra.Command{
	Use:   "gitea-cli",
	Short: "Gitea CLI - 操作自部署 Gitea 服务器的命令行工具",
	Long: `gitea-cli 是一个用于操作自部署 Gitea 服务器的命令行工具，专为 AI agent 设计。

支持仓库、Issue、PR、用户、组织、团队、Release、Webhook 等资源的完整增删查改。
凭据通过 git 自身的密钥箱（credential helper）机制管理，安全可靠。

全局标志：
  --json    所有命令以 JSON 格式输出（含错误），便于 AI 程序化解析。
            推荐 AI agent 始终携带此标志。

典型用法：
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
	rootCmd.PersistentFlags().Bool("json", false, "以 JSON 格式输出（含错误信息），便于 AI 解析")
}
