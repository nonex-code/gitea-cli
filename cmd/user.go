package cmd

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Example: `  gitea-cli user info --json       # current user
  gitea-cli user search alice`,
}

var userInfoCmd = &cobra.Command{
	Use:   "info [username]",
	Short: "Show user info (defaults to the current user)",
	Example: `  gitea-cli user info --json
  gitea-cli user info alice`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		username := ""
		if len(args) > 0 {
			username = args[0]
		}

		var u *gitea.User
		var err error
		if username == "" {
			u, _, err = cli.GetMyUserInfo()
		} else {
			u, _, err = cli.GetUserInfo(username)
		}
		if err != nil {
			fail("failed to get user info: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(u)
			return
		}

		fmt.Printf("Username:    %s\n", u.UserName)
		fmt.Printf("Full name:   %s\n", u.FullName)
		fmt.Printf("Email:       %s\n", u.Email)
		fmt.Printf("Website:     %s\n", u.Website)
		fmt.Printf("Location:    %s\n", u.Location)
		fmt.Printf("Bio:         %s\n", u.Description)
	},
}

var userSearchCmd = &cobra.Command{
	Use:     "search <keyword>",
	Short:   "Search users",
	Example: `  gitea-cli user search alice --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		users, _, err := cli.SearchUsers(gitea.SearchUsersOption{
			KeyWord: args[0],
		})
		if err != nil {
			fail("failed to search users: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(users)
			return
		}

		for _, u := range users {
			fmt.Printf("%s\t%s\n", u.UserName, u.FullName)
		}
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userInfoCmd)
	userCmd.AddCommand(userSearchCmd)

}
