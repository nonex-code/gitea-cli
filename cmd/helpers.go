package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"gitea-cli/client"
)

// mustClient creates a client and exits on error.
func mustClient() *client.Client {
	cfg := loadConfig()
	cli, err := client.New(cfg)
	if err != nil {
		fail("创建客户端失败: %v", err)
	}
	return cli
}

// splitRepo splits "owner/repo" into owner and repo.
func splitRepo(s string) (owner, repo string) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		fail("仓库格式应为 <owner>/<repo>，实际为 %q", s)
	}
	return parts[0], parts[1]
}

// printJSON outputs any value to stdout in JSON format.
func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fail("JSON 序列化失败: %v", err)
	}
	fmt.Println(string(b))
}

// printSuccess outputs a success result.
// In --json mode it outputs {"ok": true, ...extra}; otherwise it prints a
// human-readable message.
func printSuccess(message string, extra map[string]interface{}) {
	if jsonEnabled() {
		obj := map[string]interface{}{"ok": true}
		for k, v := range extra {
			obj[k] = v
		}
		printJSON(obj)
		return
	}
	fmt.Println(message)
}

// fail prints an error to stderr and exits with a non-zero exit code.
// When the global --json flag is enabled, the error is also emitted as JSON
// for easy parsing by AI agents.
func fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if jsonEnabled() {
		// Structured error output
		errObj := map[string]interface{}{
			"ok":    false,
			"error": msg,
		}
		b, _ := json.Marshal(errObj)
		fmt.Fprintln(os.Stderr, string(b))
	} else {
		fmt.Fprintf(os.Stderr, "错误: %s\n", msg)
	}
	os.Exit(1)
}

// jsonEnabled checks whether the global --json flag is enabled.
func jsonEnabled() bool {
	for _, arg := range os.Args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

// termIsTTY reports whether stdin is an interactive terminal.
func termIsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// readAllStdin reads the entire contents of stdin.
func readAllStdin() (string, error) {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
