package cmd

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// readPassword reads a password from the terminal without echoing it.
func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// Non-terminal environment (e.g. pipe): fall back to normal read
	var s string
	_, err := fmt.Scanln(&s)
	return s, err
}
