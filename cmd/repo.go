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
	Short:   "Manage repositories",
	Aliases: []string{"repository"},
	Example: `  gitea-cli repo list                 # list my repositories
  gitea-cli repo info owner/repo       # show repository details
  gitea-cli repo create my-repo        # create a repository
  gitea-cli repo star owner/repo       # star a repository`,
}

var repoListCmd = &cobra.Command{
	Use:     "list [owner]",
	Short:   "List repositories (defaults to the current user's)",
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
			fail("failed to list repositories: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(repos)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "Repository\tStars\tDescription")
		for _, r := range repos {
			fmt.Fprintf(w, "%s/%s\t%d\t%s\n", r.Owner.UserName, r.Name, r.Stars, r.Description)
		}
		w.Flush()
	},
}

var repoInfoCmd = &cobra.Command{
	Use:     "info <owner>/<repo>",
	Short:   "Show repository details",
	Example: `  gitea-cli repo info owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		r, _, err := cli.GetRepo(owner, repo)
		if err != nil {
			fail("failed to get repository info: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(r)
			return
		}

		fmt.Printf("Repository: %s\n", r.FullName)
		fmt.Printf("Description: %s\n", r.Description)
		fmt.Printf("Default branch: %s\n", r.DefaultBranch)
		fmt.Printf("Stars:      %d\n", r.Stars)
		fmt.Printf("Forks:      %d\n", r.Forks)
		fmt.Printf("Open issues: %d\n", r.OpenIssues)
		fmt.Printf("Private:    %t\n", r.Private)
		fmt.Printf("HTML URL:   %s\n", r.HTMLURL)
		fmt.Printf("Clone URL:  %s\n", r.CloneURL)
		if r.Website != "" {
			fmt.Printf("Website:    %s\n", r.Website)
		}
	},
}

var repoCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create a repository",
	Example: `  gitea-cli repo create my-repo --description "desc" --private`,
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
			fail("failed to create repository: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(repo)
			return
		}
		fmt.Printf("✓ Repository created: %s\n", repo.HTMLURL)
	},
}

var repoDeleteCmd = &cobra.Command{
	Use:     "delete <owner>/<repo>",
	Short:   "Delete a repository",
	Example: `  gitea-cli repo delete owner/repo --yes   # --yes is required in non-interactive mode`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		confirmDelete(args[0], func() {
			if _, err := cli.DeleteRepo(owner, repo); err != nil {
				fail("failed to delete repository: %v", err)
			}
			printSuccess("✓ Repository deleted: "+args[0], map[string]interface{}{"repo": args[0]})
		})
	},
}

var repoStarCmd = &cobra.Command{
	Use:     "star <owner>/<repo>",
	Short:   "Star a repository",
	Example: `  gitea-cli repo star owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.StarRepo(owner, repo); err != nil {
			fail("failed to star repository: %v", err)
		}
		printSuccess("✓ Repository starred: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoUnstarCmd = &cobra.Command{
	Use:     "unstar <owner>/<repo>",
	Short:   "Unstar a repository",
	Example: `  gitea-cli repo unstar owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.UnStarRepo(owner, repo); err != nil {
			fail("failed to unstar repository: %v", err)
		}
		printSuccess("✓ Repository unstarred: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoWatchCmd = &cobra.Command{
	Use:     "watch <owner>/<repo>",
	Short:   "Watch a repository",
	Example: `  gitea-cli repo watch owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.WatchRepo(owner, repo); err != nil {
			fail("failed to watch repository: %v", err)
		}
		printSuccess("✓ Repository watched: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoUnwatchCmd = &cobra.Command{
	Use:     "unwatch <owner>/<repo>",
	Short:   "Unwatch a repository",
	Example: `  gitea-cli repo unwatch owner/repo`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		if _, err := cli.UnWatchRepo(owner, repo); err != nil {
			fail("failed to unwatch repository: %v", err)
		}
		printSuccess("✓ Repository unwatched: "+args[0], map[string]interface{}{"repo": args[0]})
	},
}

var repoLabelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage repository labels",
}

var repoLabelListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "List repository labels",
	Example: `  gitea-cli repo label list owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		labels, _, err := cli.ListRepoLabels(owner, repo, gitea.ListLabelsOptions{})
		if err != nil {
			fail("failed to list labels: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(labels)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "Name\tColor\tDescription")
		for _, l := range labels {
			fmt.Fprintf(w, "%s\t%s\t%s\n", l.Name, l.Color, l.Description)
		}
		w.Flush()
	},
}

var repoLabelCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <name> <color>",
	Short:   "Create a repository label (color in hex, e.g. #00aabb)",
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
			fail("failed to create label: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(label)
			return
		}
		fmt.Printf("✓ Label created: %s (#%s)\n", label.Name, label.Color)
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

	repoCreateCmd.Flags().String("description", "", "repository description")
	repoCreateCmd.Flags().Bool("private", false, "make the repository private")
	repoDeleteCmd.Flags().Bool("yes", false, "skip confirmation and delete directly")

	repoLabelCreateCmd.Flags().String("description", "", "label description")
}
