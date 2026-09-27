package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var orgTeamCmd = &cobra.Command{
	Use:   "team",
	Short: "组织团队管理",
	Example: `  gitea-cli org team list orgname
  gitea-cli org team create orgname devteam --permission write`,
}

var orgTeamListCmd = &cobra.Command{
	Use:     "list <组织名>",
	Short:   "列出组织团队",
	Example: `  gitea-cli org team list orgname --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		teams, _, err := cli.ListOrgTeams(args[0], gitea.ListTeamsOptions{})
		if err != nil {
			fail("获取团队列表失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(teams)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\t名称\t描述\t权限")
		for _, t := range teams {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Name, t.Description, t.Permission)
		}
		w.Flush()
	},
}

var orgTeamCreateCmd = &cobra.Command{
	Use:     "create <组织名> <团队名>",
	Short:   "创建组织团队",
	Example: `  gitea-cli org team create orgname devteam --permission write`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		description, _ := cmd.Flags().GetString("description")
		permission, _ := cmd.Flags().GetString("permission")

		opt := gitea.CreateTeamOption{
			Name:        args[1],
			Description: description,
			Permission:  gitea.AccessMode(permission),
		}
		if opt.Permission == "" {
			opt.Permission = gitea.AccessModeRead
		}

		team, _, err := cli.CreateTeam(args[0], opt)
		if err != nil {
			fail("创建团队失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(team)
			return
		}
		fmt.Printf("✓ 团队已创建: %s (ID %d)\n", team.Name, team.ID)
	},
}

var orgTeamMemberCmd = &cobra.Command{
	Use:   "member",
	Short: "团队成员管理",
	Example: `  gitea-cli org team member list <team-id>
  gitea-cli org team member add <team-id> username`,
}

var orgTeamMemberListCmd = &cobra.Command{
	Use:     "list <团队ID>",
	Short:   "列出团队成员",
	Example: `  gitea-cli org team member list 6 --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fail("无效的团队 ID: %s", args[0])
		}

		members, _, err := cli.ListTeamMembers(id, gitea.ListTeamMembersOptions{})
		if err != nil {
			fail("获取团队成员失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(members)
			return
		}

		for _, m := range members {
			fmt.Printf("%s\t%s\n", m.UserName, m.FullName)
		}
	},
}

var orgTeamMemberAddCmd = &cobra.Command{
	Use:     "add <团队ID> <用户名>",
	Short:   "添加团队成员",
	Example: `  gitea-cli org team member add 6 username`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fail("无效的团队 ID: %s", args[0])
		}

		if _, err := cli.AddTeamMember(id, args[1]); err != nil {
			fail("添加团队成员失败: %v", err)
		}
		printSuccess(fmt.Sprintf("✓ 已添加成员 %s 到团队 %d", args[1], id), map[string]interface{}{"user": args[1], "team_id": id})
	},
}

var orgTeamMemberRemoveCmd = &cobra.Command{
	Use:     "remove <团队ID> <用户名>",
	Short:   "移除团队成员",
	Example: `  gitea-cli org team member remove 6 username`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fail("无效的团队 ID: %s", args[0])
		}

		if _, err := cli.RemoveTeamMember(id, args[1]); err != nil {
			fail("移除团队成员失败: %v", err)
		}
		printSuccess("✓ 已移除成员 "+args[1], map[string]interface{}{"user": args[1]})
	},
}

func init() {
	orgTeamCmd.AddCommand(orgTeamListCmd)
	orgTeamCmd.AddCommand(orgTeamCreateCmd)
	orgTeamCmd.AddCommand(orgTeamMemberCmd)

	orgTeamMemberCmd.AddCommand(orgTeamMemberListCmd)
	orgTeamMemberCmd.AddCommand(orgTeamMemberAddCmd)
	orgTeamMemberCmd.AddCommand(orgTeamMemberRemoveCmd)

	orgTeamCreateCmd.Flags().String("description", "", "团队描述")
	orgTeamCreateCmd.Flags().String("permission", "read", "权限 (read/write/admin)")
}
