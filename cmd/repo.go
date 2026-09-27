package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "仓库管理",
	Aliases: []string{"repository"},
	Example: `  gitea-cli repo list                 # 列出我的仓库
  gitea-cli repo info owner/repo       # 查看仓库详情
  gitea-cli repo create my-repo        # 创建仓库
  gitea-cli repo star owner/repo       # 星标仓库`,
}

var repoListCmd = &cobra.Command{
	Use:     "list [owner]",
	Short:   "列出仓库（默认列出当前用户的仓库）",
	Example: `  gitea-cli repo list --json       # list my repositories as JSON`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		var repos []*gitea.Repository
		var err error
		if len(args) > 0 {
			repos, _, err = cli.ListUserRepos(args[0], gitea.ListReposOptions{})
		} else {
			repos, _, err = cli.ListMyRepos(gitea.ListReposOptions{})
		}
		if err != nil {
			fail("获取仓库列表失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(repos)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "仓库\t星标数\t描述")
		for _, r := range repos {
			fmt.Fprintf(w, "%s/%s\t%d\t%s\n", r.Owner.UserName, r.Name, r.Stars, r.Description)
		}
		w.Flush()
	},
}

var repoInfoCmd = &cobra.Command{
	Use:     "info <owner>/<repo>",
	Short:   "查看仓库详情",
	Example: `  gitea-cli repo info owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		r, _, err := cli.GetRepo(owner, repo)
		if err != nil {
			fail("获取仓库信息失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(r)
			return
		}

		fmt.Printf("仓库:      %s\n", r.FullName)
		fmt.Printf("描述:      %s\n", r.Description)
		fmt.Printf("默认分支:  %s\n", r.DefaultBranch)
		fmt.Printf("星标数:    %d\n", r.Stars)
		fmt.Printf("Fork 数:   %d\n", r.Forks)
		fmt.Printf("开放 Issue:%d\n", r.OpenIssues)
		fmt.Printf("私有:      %t\n", r.Private)
		fmt.Printf("主页:      %s\n", r.HTMLURL)
		fmt.Printf("克隆地址:  %s\n", r.CloneURL)
		if r.Website != "" {
			fmt.Printf("网站:      %s\n", r.Website)
		}
	},
}

var repoCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "创建仓库",
	Example: `  gitea-cli repo create my-repo --description "描述" --private`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		description, _ := cmd.Flags().GetString("description")
		private, _ := cmd.Flags().GetBool("private")

		repo, _, err := cli.CreateRepo(gitea.CreateRepoOption{
			Name:        args[0],
			Description: description,
			Private:     private,
			AutoInit:    true,
		})
		if err != nil {
			fail("创建仓库失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(repo)
			return
		}
		fmt.Printf("✓ 仓库已创建: %s\n", repo.HTMLURL)
	},
}

var repoDeleteCmd = &cobra.Command{
	Use:     "delete <owner>/<repo>",
	Short:   "删除仓库",
	Example: `  gitea-cli repo delete owner/repo --yes   # --yes is required in non-interactive mode`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		confirmDelete(args[0], func() {
			if _, err := cli.DeleteRepo(owner, repo); err != nil {
				fail("删除仓库失败: %v", err)
			}
			printSuccess("✓ 仓库已删除: "+args[0], map[string]interface{}{"repo": args[0]})
		})
	},
}

var repoStarCmd = &cobra.Command{
	Use:     "star <owner>/<repo>",
	Short:   "星标仓库",
	Example: `  gitea-cli repo star owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.StarRepo(owner, repo); err != nil {
			fail("星标仓库失败: %v", err)
		}
		printSuccess("✓ 已星标仓库: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoUnstarCmd = &cobra.Command{
	Use:     "unstar <owner>/<repo>",
	Short:   "取消星标仓库",
	Example: `  gitea-cli repo unstar owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.UnStarRepo(owner, repo); err != nil {
			fail("取消星标失败: %v", err)
		}
		printSuccess("✓ 已取消星标: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoWatchCmd = &cobra.Command{
	Use:     "watch <owner>/<repo>",
	Short:   "关注仓库",
	Example: `  gitea-cli repo watch owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.WatchRepo(owner, repo); err != nil {
			fail("关注仓库失败: %v", err)
		}
		printSuccess("✓ 已关注仓库: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoUnwatchCmd = &cobra.Command{
	Use:     "unwatch <owner>/<repo>",
	Short:   "取消关注仓库",
	Example: `  gitea-cli repo unwatch owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.UnWatchRepo(owner, repo); err != nil {
			fail("取消关注失败: %v", err)
		}
		printSuccess("✓ 已取消关注: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoLabelCmd = &cobra.Command{
	Use:   "label",
	Short: "仓库标签管理",
}

var repoLabelListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "列出仓库标签",
	Example: `  gitea-cli repo label list owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		labels, _, err := cli.ListRepoLabels(owner, repo, gitea.ListLabelsOptions{})
		if err != nil {
			fail("获取标签列表失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(labels)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "名称\t颜色\t描述")
		for _, l := range labels {
			fmt.Fprintf(w, "%s\t%s\t%s\n", l.Name, l.Color, l.Description)
		}
		w.Flush()
	},
}

var repoLabelCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <名称> <颜色>",
	Short:   "创建仓库标签（颜色为十六进制，如 #00aabb）",
	Example: `  gitea-cli repo label create owner/repo "bug" "#ff0000" --description "bug label"`,
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		description, _ := cmd.Flags().GetString("description")

		label, _, err := cli.CreateLabel(owner, repo, gitea.CreateLabelOption{
			Name:        args[1],
			Color:       args[2],
			Description: description,
		})
		if err != nil {
			fail("创建标签失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(label)
			return
		}
		fmt.Printf("✓ 标签已创建: %s (#%s)\n", label.Name, label.Color)
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoInfoCmd)
	repoCmd.AddCommand(repoCreateCmd)
	repoCmd.AddCommand(repoDeleteCmd)
	repoCmd.AddCommand(repoStarCmd)
	repoCmd.AddCommand(repoUnstarCmd)
	repoCmd.AddCommand(repoWatchCmd)
	repoCmd.AddCommand(repoUnwatchCmd)
	repoCmd.AddCommand(repoLabelCmd)

	repoLabelCmd.AddCommand(repoLabelListCmd)
	repoLabelCmd.AddCommand(repoLabelCreateCmd)

	repoCreateCmd.Flags().String("description", "", "仓库描述")
	repoCreateCmd.Flags().Bool("private", false, "是否私有仓库")
	repoDeleteCmd.Flags().Bool("yes", false, "跳过确认直接删除")

	repoLabelCreateCmd.Flags().String("description", "", "标签描述")
}
