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
	Short: "Manage organizations",
	Example: `  gitea-cli org list --json
  gitea-cli org info orgname
  gitea-cli org create neworg`,
}

var orgListCmd = &cobra.Command{
	Use:     "list [user]",
	Short:   "List organizations (defaults to the current user's)",
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
			fail("failed to list organizations: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(orgs)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "Organization\tFull name\tVisibility\tDescription")
		for _, o := range orgs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", o.Name, o.FullName, o.Visibility, o.Description)
		}
		w.Flush()
	},
}

var orgInfoCmd = &cobra.Command{
	Use:     "info <org-name>",
	Short:   "Show organization info",
	Example: `  gitea-cli org info orgname --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		o, _, err := cli.GetOrg(args[0])
		if err != nil {
			fail("failed to get organization info: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(o)
			return
		}

		fmt.Printf("Name:       %s\n", o.Name)
		fmt.Printf("Full name:  %s\n", o.FullName)
		fmt.Printf("Visibility: %s\n", o.Visibility)
		fmt.Printf("Description: %s\n", o.Description)
		fmt.Printf("Website:    %s\n", o.Website)
		fmt.Printf("Location:   %s\n", o.Location)
	},
}

var orgCreateCmd = &cobra.Command{
	Use:     "create <org-name>",
	Short:   "Create an organization",
	Example: `  gitea-cli org create neworg --description "desc"`,
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
			fail("failed to create organization: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(o)
			return
		}
		fmt.Printf("✓ Organization created: %s\n", o.Name)
	},
}

var orgDeleteCmd = &cobra.Command{
	Use:     "delete <org-name>",
	Short:   "Delete an organization",
	Example: `  gitea-cli org delete orgname --yes`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		confirmDelete(fmt.Sprintf("organization %s", args[0]), func() {
			if _, err := cli.DeleteOrg(args[0]); err != nil {
				fail("failed to delete organization: %v", err)
			}
			printSuccess("✓ Organization "+args[0]+" deleted", map[string]interface{}{"org": args[0]})
		})
	},
}

var orgMemberCmd = &cobra.Command{
	Use:   "member",
	Short: "Manage organization members",
	Example: `  gitea-cli org member list orgname
  gitea-cli org member add orgname username --team Owners`,
}

var orgMemberListCmd = &cobra.Command{
	Use:     "list <org-name>",
	Short:   "List organization members",
	Example: `  gitea-cli org member list orgname --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		members, _, err := cli.ListOrgMembership(args[0], gitea.ListOrgMembershipOption{})
		if err != nil {
			fail("failed to list organization members: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(members)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "Username\tFull name\tEmail")
		for _, m := range members {
			fmt.Fprintf(w, "%s\t%s\t%s\n", m.UserName, m.FullName, m.Email)
		}
		w.Flush()
	},
}

var orgMemberAddCmd = &cobra.Command{
	Use:     "add <org-name> <username>",
	Short:   "Add an organization member (via a team)",
	Long:    "In Gitea, organization members must be added via a team. Create a team first if none exists.\nUse --team to specify the team name (default: Owners).",
	Example: `  gitea-cli org member add orgname username --team Owners`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		teamName, _ := cmd.Flags().GetString("team")

		// Find the team
		teams, _, err := cli.SearchOrgTeams(args[0], &gitea.SearchTeamsOptions{})
		if err != nil {
			fail("failed to list teams: %v", err)
		}

		var teamID int64
		for _, t := range teams {
			if t.Name == teamName {
				teamID = t.ID
				break
			}
		}
		if teamID == 0 {
			fail("team %q not found; create it or check the team name", teamName)
		}

		if _, err := cli.AddTeamMember(teamID, args[1]); err != nil {
			fail("failed to add member: %v", err)
		}
		printSuccess(fmt.Sprintf("✓ Added member %s to team %s", args[1], teamName), map[string]interface{}{"user": args[1], "team": teamName})
	},
}

var orgMemberRemoveCmd = &cobra.Command{
	Use:     "remove <org-name> <username>",
	Short:   "Remove an organization member",
	Example: `  gitea-cli org member remove orgname username --yes`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		confirmDelete(fmt.Sprintf("member %s (organization %s)", args[1], args[0]), func() {
			if _, err := cli.DeleteOrgMembership(args[0], args[1]); err != nil {
				fail("failed to remove member: %v", err)
			}
			printSuccess("✓ Removed member "+args[1], map[string]interface{}{"user": args[1]})
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

	orgCreateCmd.Flags().String("fullname", "", "organization full name")
	orgCreateCmd.Flags().String("description", "", "organization description")
	orgCreateCmd.Flags().String("visibility", "", "visibility (public/limited/private)")
	orgDeleteCmd.Flags().Bool("yes", false, "skip confirmation and delete directly")

	orgMemberAddCmd.Flags().String("team", "Owners", "team name (used when adding a member)")
	orgMemberRemoveCmd.Flags().Bool("yes", false, "skip confirmation and delete directly")
}
