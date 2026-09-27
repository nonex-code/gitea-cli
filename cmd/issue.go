package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues",
	Example: `  gitea-cli issue list owner/repo --json
  gitea-cli issue info owner/repo 1
  gitea-cli issue create owner/repo "title" --body "content"`,
}

var issueListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "List issues in a repository",
	Example: `  gitea-cli issue list owner/repo --state open --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		state, _ := cmd.Flags().GetString("state")

		issues, _, err := cli.ListRepoIssues(owner, repo, gitea.ListIssueOption{
			State: gitea.StateType(state),
		})
		if err != nil {
			fail("failed to list issues: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(issues)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "#\tTitle\tState\tAuthor")
		for _, i := range issues {
			fmt.Fprintf(w, "#%d\t%s\t%s\t%s\n", i.Index, i.Title, i.State, i.Poster.UserName)
		}
		w.Flush()
	},
}

var issueInfoCmd = &cobra.Command{
	Use:     "info <owner>/<repo> <number>",
	Short:   "Show issue details",
	Example: `  gitea-cli issue info owner/repo 1 --json`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid issue number: %s", args[1])
		}

		issue, _, err := cli.GetIssue(owner, repo, index)
		if err != nil {
			fail("failed to get issue: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(issue)
			return
		}

		fmt.Printf("#%d %s\n", issue.Index, issue.Title)
		fmt.Printf("State:     %s\n", issue.State)
		fmt.Printf("Author:    %s\n", issue.Poster.UserName)
		if len(issue.Assignees) > 0 {
			names := make([]string, 0, len(issue.Assignees))
			for _, a := range issue.Assignees {
				names = append(names, a.UserName)
			}
			fmt.Printf("Assignees: %v\n", names)
		}
		fmt.Printf("Created:   %s\n", issue.Created.Format("2006-01-02 15:04:05"))
		fmt.Printf("URL:       %s\n", issue.HTMLURL)
		if issue.Body != "" {
			fmt.Printf("\n%s\n", issue.Body)
		}
	},
}

var issueCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <title>",
	Short:   "Create an issue",
	Example: `  gitea-cli issue create owner/repo "title" --body "content"`,
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		body, _ := cmd.Flags().GetString("body")

		title := ""
		for i := 1; i < len(args); i++ {
			if i > 1 {
				title += " "
			}
			title += args[i]
		}

		issue, _, err := cli.CreateIssue(owner, repo, gitea.CreateIssueOption{
			Title: title,
			Body:  body,
		})
		if err != nil {
			fail("failed to create issue: %v", err)
		}
		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(issue)
			return
		}
		fmt.Printf("✓ Issue created: #%d %s\n", issue.Index, issue.Title)
	},
}

var issueCloseCmd = &cobra.Command{
	Use:     "close <owner>/<repo> <number>",
	Short:   "Close an issue",
	Example: `  gitea-cli issue close owner/repo 1`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid issue number: %s", args[1])
		}

		state := gitea.StateClosed
		_, _, err = cli.EditIssue(owner, repo, index, gitea.EditIssueOption{
			State: &state,
		})
		if err != nil {
			fail("failed to close issue: %v", err)
		}
		printSuccess(fmt.Sprintf("✓ Issue #%d closed", index), map[string]interface{}{"index": index, "state": "closed"})
	},
}

var issueReopenCmd = &cobra.Command{
	Use:     "reopen <owner>/<repo> <number>",
	Short:   "Reopen an issue",
	Example: `  gitea-cli issue reopen owner/repo 1`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid issue number: %s", args[1])
		}

		state := gitea.StateOpen
		_, _, err = cli.EditIssue(owner, repo, index, gitea.EditIssueOption{
			State: &state,
		})
		if err != nil {
			fail("failed to reopen issue: %v", err)
		}
		printSuccess(fmt.Sprintf("✓ Issue #%d reopened", index), map[string]interface{}{"index": index, "state": "open"})
	},
}

var issueDeleteCmd = &cobra.Command{
	Use:     "delete <owner>/<repo> <number>",
	Short:   "Delete an issue",
	Example: `  gitea-cli issue delete owner/repo 1 --yes`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid issue number: %s", args[1])
		}

		confirmDelete(fmt.Sprintf("Issue #%d (%s)", index, args[0]), func() {
			if _, err := cli.DeleteIssue(owner, repo, index); err != nil {
				fail("failed to delete issue: %v", err)
			}
			printSuccess(fmt.Sprintf("✓ Issue #%d deleted", index), map[string]interface{}{"index": index})
		})
	},
}

var issueCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage issue comments",
	Example: `  gitea-cli issue comment list owner/repo 1
  gitea-cli issue comment add owner/repo 1 "comment text"`,
}

var issueCommentListCmd = &cobra.Command{
	Use:     "list <owner>/<repo> <number>",
	Short:   "List comments on an issue",
	Example: `  gitea-cli issue comment list owner/repo 1 --json`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid issue number: %s", args[1])
		}

		comments, _, err := cli.ListIssueComments(owner, repo, index, gitea.ListIssueCommentOptions{})
		if err != nil {
			fail("failed to list comments: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(comments)
			return
		}

		for _, c := range comments {
			fmt.Printf("--- %s (%s) ---\n", c.Poster.UserName, c.Created.Format("2006-01-02 15:04"))
			fmt.Printf("%s\n", c.Body)
		}
	},
}

var issueCommentAddCmd = &cobra.Command{
	Use:     "add <owner>/<repo> <number> <comment>",
	Short:   "Add a comment to an issue",
	Example: `  gitea-cli issue comment add owner/repo 1 "comment text"`,
	Args:    cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		index, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid issue number: %s", args[1])
		}

		body := ""
		for i := 2; i < len(args); i++ {
			if i > 2 {
				body += " "
			}
			body += args[i]
		}

		comment, _, err := cli.CreateIssueComment(owner, repo, index, gitea.CreateIssueCommentOption{
			Body: body,
		})
		if err != nil {
			fail("failed to add comment: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(comment)
			return
		}
		fmt.Printf("✓ Comment added (ID %d)\n", comment.ID)
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)
	issueCmd.AddCommand(issueListCmd)
	issueCmd.AddCommand(issueInfoCmd)
	issueCmd.AddCommand(issueCreateCmd)
	issueCmd.AddCommand(issueCloseCmd)
	issueCmd.AddCommand(issueReopenCmd)
	issueCmd.AddCommand(issueDeleteCmd)
	issueCmd.AddCommand(issueCommentCmd)

	issueCommentCmd.AddCommand(issueCommentListCmd)
	issueCommentCmd.AddCommand(issueCommentAddCmd)

	issueListCmd.Flags().String("state", "open", "filter by state (open/closed/all)")
	issueCreateCmd.Flags().String("body", "", "issue body")
	issueDeleteCmd.Flags().Bool("yes", false, "skip confirmation and delete directly")

}
