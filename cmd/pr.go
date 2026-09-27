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
	Short: "Manage pull requests",
	Example: `  gitea-cli pr list owner/repo --json
  gitea-cli pr info owner/repo 3
  gitea-cli pr create owner/repo "title" --head feature --base main`,
}

var prListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "List pull requests in a repository",
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
			fail("failed to list pull requests: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(prs)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "#\tTitle\tState\tAuthor")
		for _, p := range prs {
			fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n", p.Index, p.Title, p.State, p.Poster.UserName)
		}
		w.Flush()
	},
}

var prInfoCmd = &cobra.Command{
	Use:     "info <owner>/<repo> <number>",
	Short:   "Show pull request details",
	Example: `  gitea-cli pr info owner/repo 3 --json`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid PR number: %s", args[1])
		}

		pr, _, err := cli.GetPullRequest(owner, repo, index)
		if err != nil {
			fail("failed to get pull request: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(pr)
			return
		}

		fmt.Printf("#%d %s\n", pr.Index, pr.Title)
		fmt.Printf("State:      %s\n", pr.State)
		fmt.Printf("Author:     %s\n", pr.Poster.UserName)
		fmt.Printf("Head branch: %s\n", pr.Head.Name)
		fmt.Printf("Base branch: %s\n", pr.Base.Name)
		fmt.Printf("Mergeable:  %t\n", pr.Mergeable)
		fmt.Printf("URL:      %s\n", pr.HTMLURL)
		if pr.Body != "" {
			fmt.Printf("\n%s\n", pr.Body)
		}
	},
}

var prCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <title>",
	Short:   "Create a pull request",
	Example: `  gitea-cli pr create owner/repo "title" --head feature --base main`,
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
			fmt.Fprintln(os.Stderr, "error: --head branch is required")
			os.Exit(1)
		}

		pr, _, err := cli.CreatePullRequest(owner, repo, gitea.CreatePullRequestOption{
			Title: title,
			Body:  body,
			Head:  head,
			Base:  base,
		})
		if err != nil {
			fail("failed to create pull request: %v", err)
		}
		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(pr)
			return
		}
		fmt.Printf("✓ PR created: #%d %s\n", pr.Index, pr.Title)
	},
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prInfoCmd)
	prCmd.AddCommand(prCreateCmd)

	prListCmd.Flags().String("state", "open", "filter by state (open/closed/all)")
	prCreateCmd.Flags().String("head", "", "head branch")
	prCreateCmd.Flags().String("base", "", "base branch (defaults to the repository default branch)")
	prCreateCmd.Flags().String("body", "", "PR body")
}
