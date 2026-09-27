package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// confirmDelete performs a confirmation before deletion.
// Rules:
//   - With --yes/-y flag: execute directly without asking.
//   - With interactive stdin: prompt the user to type yes to confirm.
//   - With non-interactive stdin (e.g. AI agent pipe): refuse without --yes and
//     error out instead of blocking on input. This is the key AI-agent-friendly
//     behavior.
func confirmDelete(name string, do func()) {
	if hasFlag("--yes") || hasFlag("-y") {
		do()
		return
	}

	// Non-interactive environment: refuse and ask for --yes.
	if !termIsTTY() {
		fail("deleting %s requires confirmation, but this is a non-interactive environment; add --yes to confirm", name)
	}

	fmt.Printf("确认删除 %s？此操作不可恢复。输入 yes 确认: ", name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input != "yes" && input != "y" {
		fmt.Println("已取消")
		os.Exit(0)
	}
	do()
}

// hasFlag checks whether the given flag appears in the command-line arguments.
func hasFlag(flag string) bool {
	for _, arg := range os.Args {
		if arg == flag {
			return true
		}
	}
	return false
}
