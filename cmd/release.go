package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release 管理",
	Example: `  gitea-cli release list owner/repo --json
  gitea-cli release create owner/repo v1.0.0 --note "发布说明"`,
}

var releaseListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "列出仓库的 Release",
	Example: `  gitea-cli release list owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		releases, _, err := cli.ListReleases(owner, repo, gitea.ListReleasesOptions{})
		if err != nil {
			fail("获取 Release 列表失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(releases)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "标签\t标题\t草稿\t预发布\t发布时间")
		for _, r := range releases {
			fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%s\n", r.TagName, r.Title, r.IsDraft, r.IsPrerelease, r.PublishedAt.Format("2006-01-02 15:04"))
		}
		w.Flush()
	},
}

var releaseCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <标签>",
	Short:   "创建 Release",
	Example: `  gitea-cli release create owner/repo v1.0.0 --note "发布说明"`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		opt := gitea.CreateReleaseOption{
			TagName: args[1],
		}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			opt.Title = v
		} else {
			opt.Title = args[1]
		}
		if v, _ := cmd.Flags().GetString("note"); v != "" {
			opt.Note = v
		}
		if v, _ := cmd.Flags().GetString("target"); v != "" {
			opt.Target = v
		}
		opt.IsDraft, _ = cmd.Flags().GetBool("draft")
		opt.IsPrerelease, _ = cmd.Flags().GetBool("prerelease")

		r, _, err := cli.CreateRelease(owner, repo, opt)
		if err != nil {
			fail("创建 Release 失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(r)
			return
		}
		fmt.Printf("✓ Release 已创建: %s (%s)\n", r.Title, r.HTMLURL)
	},
}

var releaseDeleteCmd = &cobra.Command{
	Use:     "delete <owner>/<repo> <标签>",
	Short:   "删除 Release（按标签）",
	Example: `  gitea-cli release delete owner/repo v1.0.0 --yes`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		confirmDelete(fmt.Sprintf("Release %s (%s)", args[1], args[0]), func() {
			if _, err := cli.DeleteReleaseByTag(owner, repo, args[1]); err != nil {
				fail("删除 Release 失败: %v", err)
			}
			printSuccess(fmt.Sprintf("✓ Release %s 已删除", args[1]), map[string]interface{}{"tag": args[1]})
		})
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.AddCommand(releaseListCmd)
	releaseCmd.AddCommand(releaseCreateCmd)
	releaseCmd.AddCommand(releaseDeleteCmd)

	releaseCreateCmd.Flags().String("title", "", "标题（默认使用标签名）")
	releaseCreateCmd.Flags().String("note", "", "发布说明")
	releaseCreateCmd.Flags().String("target", "", "目标分支/提交")
	releaseCreateCmd.Flags().Bool("draft", false, "是否草稿")
	releaseCreateCmd.Flags().Bool("prerelease", false, "是否预发布")
	releaseDeleteCmd.Flags().Bool("yes", false, "跳过确认直接删除")
}
