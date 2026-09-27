package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Pull Request 管理",
	Example: `  gitea-cli pr list owner/repo --json
  gitea-cli pr info owner/repo 3
  gitea-cli pr create owner/repo "标题" --head feature --base main`,
}

var prListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "列出仓库的 PR",
	Example: `  gitea-cli pr list owner/repo --state open --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		state, _ := cmd.Flags().GetString("state")

		prs, _, err := cli.ListRepoPullRequests(owner, repo, gitea.ListPullRequestsOptions{
			State: gitea.StateType(state),
		})
		if err != nil {
			fail("获取 PR 列表失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(prs)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "#\t标题\t状态\t作者")
		for _, p := range prs {
			fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n", p.Index, p.Title, p.State, p.Poster.UserName)
		}
		w.Flush()
	},
}

var prInfoCmd = &cobra.Command{
	Use:     "info <owner>/<repo> <编号>",
	Short:   "查看 PR 详情",
	Example: `  gitea-cli pr info owner/repo 3 --json`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("无效的 PR 编号: %s", args[1])
		}

		pr, _, err := cli.GetPullRequest(owner, repo, index)
		if err != nil {
			fail("获取 PR 失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(pr)
			return
		}

		fmt.Printf("#%d %s\n", pr.Index, pr.Title)
		fmt.Printf("状态:     %s\n", pr.State)
		fmt.Printf("作者:     %s\n", pr.Poster.UserName)
		fmt.Printf("源分支:   %s\n", pr.Head.Name)
		fmt.Printf("目标分支: %s\n", pr.Base.Name)
		fmt.Printf("可合并:   %t\n", pr.Mergeable)
		fmt.Printf("URL:      %s\n", pr.HTMLURL)
		if pr.Body != "" {
			fmt.Printf("\n%s\n", pr.Body)
		}
	},
}

var prCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <标题>",
	Short:   "创建 PR",
	Example: `  gitea-cli pr create owner/repo "标题" --head feature --base main`,
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		head, _ := cmd.Flags().GetString("head")
		base, _ := cmd.Flags().GetString("base")
		body, _ := cmd.Flags().GetString("body")

		title := ""
		for i := 1; i < len(args); i++ {
			if i > 1 {
				title += " "
			}
			title += args[i]
		}

		if head == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --head 分支")
			os.Exit(1)
		}

		pr, _, err := cli.CreatePullRequest(owner, repo, gitea.CreatePullRequestOption{
			Title: title,
			Body:  body,
			Head:  head,
			Base:  base,
		})
		if err != nil {
			fail("创建 PR 失败: %v", err)
		}
		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(pr)
			return
		}
		fmt.Printf("✓ PR 已创建: #%d %s\n", pr.Index, pr.Title)
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prInfoCmd)
	prCmd.AddCommand(prCreateCmd)

	prListCmd.Flags().String("state", "open", "过滤状态 (open/closed/all)")
	prCreateCmd.Flags().String("head", "", "源分支")
	prCreateCmd.Flags().String("base", "", "目标分支（默认仓库默认分支）")
	prCreateCmd.Flags().String("body", "", "PR 内容")
}
