package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "组织管理",
	Example: `  gitea-cli org list --json
  gitea-cli org info orgname
  gitea-cli org create neworg`,
}

var orgListCmd = &cobra.Command{
	Use:     "list [user]",
	Short:   "列出组织（默认列出当前用户的组织）",
	Example: `  gitea-cli org list --json`,
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		var orgs []*gitea.Organization
		var err error
		if len(args) > 0 {
			orgs, _, err = cli.ListUserOrgs(args[0], gitea.ListOrgsOptions{})
		} else {
			orgs, _, err = cli.ListMyOrgs(gitea.ListOrgsOptions{})
		}
		if err != nil {
			fail("获取组织列表失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(orgs)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "组织\t全名\t可见性\t描述")
		for _, o := range orgs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", o.Name, o.FullName, o.Visibility, o.Description)
		}
		w.Flush()
	},
}

var orgInfoCmd = &cobra.Command{
	Use:     "info <组织名>",
	Short:   "查看组织信息",
	Example: `  gitea-cli org info orgname --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		o, _, err := cli.GetOrg(args[0])
		if err != nil {
			fail("获取组织信息失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(o)
			return
		}

		fmt.Printf("组织名: %s\n", o.Name)
		fmt.Printf("全名:   %s\n", o.FullName)
		fmt.Printf("可见性: %s\n", o.Visibility)
		fmt.Printf("描述:   %s\n", o.Description)
		fmt.Printf("网站:   %s\n", o.Website)
		fmt.Printf("位置:   %s\n", o.Location)
	},
}

var orgCreateCmd = &cobra.Command{
	Use:     "create <组织名>",
	Short:   "创建组织",
	Example: `  gitea-cli org create neworg --description "描述"`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		opt := gitea.CreateOrgOption{Name: args[0]}
		if v, _ := cmd.Flags().GetString("fullname"); v != "" {
			opt.FullName = v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			opt.Description = v
		}
		if v, _ := cmd.Flags().GetString("visibility"); v != "" {
			opt.Visibility = gitea.VisibleType(v)
		}

		o, _, err := cli.CreateOrg(opt)
		if err != nil {
			fail("创建组织失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(o)
			return
		}
		fmt.Printf("✓ 组织已创建: %s\n", o.Name)
	},
}

var orgDeleteCmd = &cobra.Command{
	Use:     "delete <组织名>",
	Short:   "删除组织",
	Example: `  gitea-cli org delete orgname --yes`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		confirmDelete(fmt.Sprintf("组织 %s", args[0]), func() {
			if _, err := cli.DeleteOrg(args[0]); err != nil {
				fail("删除组织失败: %v", err)
			}
			printSuccess("✓ 组织 "+args[0]+" 已删除", map[string]interface{}{"org": args[0]})
		})
	},
}

var orgMemberCmd = &cobra.Command{
	Use:   "member",
	Short: "组织成员管理",
	Example: `  gitea-cli org member list orgname
  gitea-cli org member add orgname username --team Owners`,
}

var orgMemberListCmd = &cobra.Command{
	Use:     "list <组织名>",
	Short:   "列出组织成员",
	Example: `  gitea-cli org member list orgname --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		members, _, err := cli.ListOrgMembership(args[0], gitea.ListOrgMembershipOption{})
		if err != nil {
			fail("获取组织成员失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(members)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "用户名\t全名\t邮箱")
		for _, m := range members {
			fmt.Fprintf(w, "%s\t%s\t%s\n", m.UserName, m.FullName, m.Email)
		}
		w.Flush()
	},
}

var orgMemberAddCmd = &cobra.Command{
	Use:     "add <组织名> <用户名>",
	Short:   "添加组织成员（通过团队）",
	Long:    "Gitea 中组织成员需通过团队（team）添加。若组织无团队，请先创建团队。\n可用 --team 指定团队名称（默认 Owners）。",
	Example: `  gitea-cli org member add orgname username --team Owners`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		teamName, _ := cmd.Flags().GetString("team")

		// 查找团队
		teams, _, err := cli.SearchOrgTeams(args[0], &gitea.SearchTeamsOptions{})
		if err != nil {
			fail("获取团队列表失败: %v", err)
		}

		var teamID int64
		for _, t := range teams {
			if t.Name == teamName {
				teamID = t.ID
				break
			}
		}
		if teamID == 0 {
			fail("未找到团队 %q，请先创建或检查团队名", teamName)
		}

		if _, err := cli.AddTeamMember(teamID, args[1]); err != nil {
			fail("添加成员失败: %v", err)
		}
		printSuccess(fmt.Sprintf("✓ 已添加成员 %s 到团队 %s", args[1], teamName), map[string]interface{}{"user": args[1], "team": teamName})
	},
}

var orgMemberRemoveCmd = &cobra.Command{
	Use:     "remove <组织名> <用户名>",
	Short:   "移除组织成员",
	Example: `  gitea-cli org member remove orgname username --yes`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		confirmDelete(fmt.Sprintf("成员 %s (组织 %s)", args[1], args[0]), func() {
			if _, err := cli.DeleteOrgMembership(args[0], args[1]); err != nil {
				fail("移除成员失败: %v", err)
			}
			printSuccess("✓ 已移除成员 "+args[1], map[string]interface{}{"user": args[1]})
		})
	},
}

func init() {
	rootCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(orgListCmd)
	orgCmd.AddCommand(orgInfoCmd)
	orgCmd.AddCommand(orgCreateCmd)
	orgCmd.AddCommand(orgDeleteCmd)
	orgCmd.AddCommand(orgMemberCmd)
	orgCmd.AddCommand(orgTeamCmd)

	orgMemberCmd.AddCommand(orgMemberListCmd)
	orgMemberCmd.AddCommand(orgMemberAddCmd)
	orgMemberCmd.AddCommand(orgMemberRemoveCmd)

	orgCreateCmd.Flags().String("fullname", "", "组织全名")
	orgCreateCmd.Flags().String("description", "", "组织描述")
	orgCreateCmd.Flags().String("visibility", "", "可见性 (public/limited/private)")
	orgDeleteCmd.Flags().Bool("yes", false, "跳过确认直接删除")

	orgMemberAddCmd.Flags().String("team", "Owners", "团队名称（用于添加成员）")
	orgMemberRemoveCmd.Flags().Bool("yes", false, "跳过确认直接删除")
}
