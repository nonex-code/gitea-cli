package cmd

import (
	"fmt"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "用户管理",
	Example: `  gitea-cli user info --json       # 当前用户
  gitea-cli user search alice`,
}

var userInfoCmd = &cobra.Command{
	Use:   "info [username]",
	Short: "查看用户信息（默认查看当前用户）",
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
			fail("获取用户信息失败: %v", err)
		}

		if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
			printJSON(u)
			return
		}

		fmt.Printf("用户名:   %s\n", u.UserName)
		fmt.Printf("全名:     %s\n", u.FullName)
		fmt.Printf("邮箱:     %s\n", u.Email)
		fmt.Printf("主页:     %s\n", u.Website)
		fmt.Printf("位置:     %s\n", u.Location)
		fmt.Printf("简介:     %s\n", u.Description)
	},
}

var userSearchCmd = &cobra.Command{
	Use:     "search <关键字>",
	Short:   "搜索用户",
	Example: `  gitea-cli user search alice --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli := mustClient()

		users, _, err := cli.SearchUsers(gitea.SearchUsersOption{
			KeyWord: args[0],
		})
		if err != nil {
			fail("搜索用户失败: %v", err)
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
