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
	Short: "Manage organization teams",
	Example: `  gitea-cli org team list orgname
  gitea-cli org team create orgname devteam --permission write`,
}

var orgTeamListCmd = &cobra.Command{
	Use:     "list <org-name>",
	Short:   "List organization teams",
	Example: `  gitea-cli org team list orgname --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		teams, _, err := cli.ListOrgTeams(args[0], gitea.ListTeamsOptions{})
		if err != nil {
			fail("failed to list teams: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(teams)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tName\tDescription\tPermission")
		for _, t := range teams {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", t.ID, t.Name, t.Description, t.Permission)
		}
		w.Flush()
	},
}

var orgTeamCreateCmd = &cobra.Command{
	Use:     "create <org-name> <team-name>",
	Short:   "Create an organization team",
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
			fail("failed to create team: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(team)
			return
		}
		fmt.Printf("✓ Team created: %s (ID %d)\n", team.Name, team.ID)
	},
}

var orgTeamMemberCmd = &cobra.Command{
	Use:   "member",
	Short: "Manage team members",
	Example: `  gitea-cli org team member list <team-id>
  gitea-cli org team member add <team-id> username`,
}

var orgTeamMemberListCmd = &cobra.Command{
	Use:     "list <team-id>",
	Short:   "List team members",
	Example: `  gitea-cli org team member list 6 --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fail("invalid team ID: %s", args[0])
		}

		members, _, err := cli.ListTeamMembers(id, gitea.ListTeamMembersOptions{})
		if err != nil {
			fail("failed to list team members: %v", err)
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
	Use:     "add <team-id> <username>",
	Short:   "Add a team member",
	Example: `  gitea-cli org team member add 6 username`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fail("invalid team ID: %s", args[0])
		}

		if _, err := cli.AddTeamMember(id, args[1]); err != nil {
			fail("failed to add team member: %v", err)
		}
		printSuccess(fmt.Sprintf("✓ Added member %s to team %d", args[1], id), map[string]interface{}{"user": args[1], "team_id": id})
	},
}

var orgTeamMemberRemoveCmd = &cobra.Command{
	Use:     "remove <team-id> <username>",
	Short:   "Remove a team member",
	Example: `  gitea-cli org team member remove 6 username`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fail("invalid team ID: %s", args[0])
		}

		if _, err := cli.RemoveTeamMember(id, args[1]); err != nil {
			fail("failed to remove team member: %v", err)
		}
		printSuccess("✓ Removed member "+args[1], map[string]interface{}{"user": args[1]})
	},
}

func init() {
	orgTeamCmd.AddCommand(orgTeamListCmd)
	orgTeamCmd.AddCommand(orgTeamCreateCmd)
	orgTeamCmd.AddCommand(orgTeamMemberCmd)

	orgTeamMemberCmd.AddCommand(orgTeamMemberListCmd)
	orgTeamMemberCmd.AddCommand(orgTeamMemberAddCmd)
	orgTeamMemberCmd.AddCommand(orgTeamMemberRemoveCmd)

	orgTeamCreateCmd.Flags().String("description", "", "team description")
	orgTeamCreateCmd.Flags().String("permission", "read", "permission (read/write/admin)")
}
