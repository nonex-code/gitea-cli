package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks",
	Example: `  gitea-cli webhook list owner/repo --json
  gitea-cli webhook create owner/repo gitea https://example.com/hook`,
}

var webhookListCmd = &cobra.Command{
	Use:     "list <owner>/<repo>",
	Short:   "List repository webhooks",
	Example: `  gitea-cli webhook list owner/repo --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		hooks, _, err := cli.ListRepoHooks(owner, repo, gitea.ListHooksOptions{})
		if err != nil {
			fail("failed to list webhooks: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(hooks)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tType\tTarget URL\tActive\tEvents")
		for _, h := range hooks {
			fmt.Fprintf(w, "%d\t%s\t%s\t%t\t%v\n", h.ID, h.Type, h.Config["url"], h.Active, h.Events)
		}
		w.Flush()
	},
}

var webhookCreateCmd = &cobra.Command{
	Use:     "create <owner>/<repo> <type> <target-url>",
	Short:   "Create a repository webhook",
	Example: `  gitea-cli webhook create owner/repo gitea https://example.com/hook --secret xxx`,
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		hookType := gitea.HookType(args[1])
		targetURL := args[2]

		config := map[string]string{
			"url":          targetURL,
			"content_type": "json",
		}
		if secret, _ := cmd.Flags().GetString("secret"); secret != "" {
			config["secret"] = secret
		}

		events, _ := cmd.Flags().GetStringArray("events")

		opt := gitea.CreateHookOption{
			Type:   hookType,
			Config: config,
			Events: events,
			Active: true,
		}

		h, _, err := cli.CreateRepoHook(owner, repo, opt)
		if err != nil {
			fail("failed to create webhook: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(h)
			return
		}
		fmt.Printf("✓ Webhook created: ID %d (%s)\n", h.ID, h.Type)
	},
}

var webhookDeleteCmd = &cobra.Command{
	Use:     "delete <owner>/<repo> <ID>",
	Short:   "Delete a repository webhook",
	Example: `  gitea-cli webhook delete owner/repo 2 --yes`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		owner, repo := splitRepo(args[0])
		cli := mustClient()

		id, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fail("invalid webhook ID: %s", args[1])
		}

		confirmDelete(fmt.Sprintf("Webhook ID %d (%s)", id, args[0]), func() {
			if _, err := cli.DeleteRepoHook(owner, repo, id); err != nil {
				fail("failed to delete webhook: %v", err)
			}
			printSuccess(fmt.Sprintf("✓ Webhook ID %d deleted", id), map[string]interface{}{"id": id})
		})
	},
}

func init() {
	rootCmd.AddCommand(webhookCmd)
	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookCreateCmd)
	webhookCmd.AddCommand(webhookDeleteCmd)

	webhookCreateCmd.Flags().String("secret", "", "webhook secret")
	webhookCreateCmd.Flags().StringArray("events", []string{"push"}, "trigger events (repeatable, e.g. --events push --events create)")
	webhookDeleteCmd.Flags().Bool("yes", false, "skip confirmation and delete directly")
}
