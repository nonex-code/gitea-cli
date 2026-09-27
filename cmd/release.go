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
	Short: "Manage releases",
	Example: `  gitea-cli release list owner/repo --json
  gitea-cli release create owner/repo v1.0.0 --note "release notes"`,
}

var releaseListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "List releases in a repository",
	Example: `  gitea-cli release list owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		releases, _, err := cli.ListReleases(owner, repo, gitea.ListReleasesOptions{})
		if err != nil {
			fail("failed to list releases: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(releases)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "Tag\tTitle\tDraft\tPrerelease\tPublished")
		for _, r := range releases {
			fmt.Fprintf(w, "%s\t%s\t%t\t%t\t%s\n", r.TagName, r.Title, r.IsDraft, r.IsPrerelease, r.PublishedAt.Format("2006-01-02 15:04"))
		}
		w.Flush()
	},
}

var releaseCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <tag>",
	Short:   "Create a release",
	Example: `  gitea-cli release create owner/repo v1.0.0 --note "release notes"`,
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
			fail("failed to create release: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(r)
			return
		}
		fmt.Printf("✓ Release created: %s (%s)\n", r.Title, r.HTMLURL)
	},
}

var releaseDeleteCmd = &cobra.Command{
	Use:     "delete <owner>/<repo> <tag>",
	Short:   "Delete a release (by tag)",
	Example: `  gitea-cli release delete owner/repo v1.0.0 --yes`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		confirmDelete(fmt.Sprintf("Release %s (%s)", args[1], args[0]), func() {
			if _, err := cli.DeleteReleaseByTag(owner, repo, args[1]); err != nil {
				fail("failed to delete release: %v", err)
			}
			printSuccess(fmt.Sprintf("✓ Release %s deleted", args[1]), map[string]interface{}{"tag": args[1]})
		})
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	releaseCmd.AddCommand(releaseListCmd)
	releaseCmd.AddCommand(releaseCreateCmd)
	releaseCmd.AddCommand(releaseDeleteCmd)

	releaseCreateCmd.Flags().String("title", "", "title (defaults to the tag name)")
	releaseCreateCmd.Flags().String("note", "", "release notes")
	releaseCreateCmd.Flags().String("target", "", "target branch/commit")
	releaseCreateCmd.Flags().Bool("draft", false, "mark as draft")
	releaseCreateCmd.Flags().Bool("prerelease", false, "mark as prerelease")
	releaseDeleteCmd.Flags().Bool("yes", false, "skip confirmation and delete directly")
}
