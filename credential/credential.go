package credential

import (
	"bufio"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// Credential represents a set of credentials.
type Credential struct {
	Protocol string
	Host     string
	Username string
	Password string
}

// Get retrieves credentials via the git credential mechanism
// (keyring / credential helper). This follows git's own secret management
// convention and supports various backends configured via credential.helper
// (e.g. libsecret, osxkeychain, pass, store, etc.).
func Get(serverURL, username string) (*Credential, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %w", err)
	}

	protocol := u.Scheme
	host := u.Host

	// Build input: protocol, host, username
	input := fmt.Sprintf("protocol=%s\nhost=%s\n", protocol, host)
	if username != "" {
		input += fmt.Sprintf("username=%s\n", username)
	}
	input += "\n"

	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git credential fill: %w", err)
	}

	cred := &Credential{Protocol: protocol, Host: host}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		switch key {
		case "username":
			cred.Username = value
		case "password":
			cred.Password = value
		}
	}

	if cred.Password == "" {
		return nil, fmt.Errorf("failed to get password from git credential; make sure credentials have been stored via git credential")
	}
	return cred, nil
}

// Store stores credentials into the keyring via the git credential mechanism.
func Store(serverURL, username, password string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("failed to parse server URL: %w", err)
	}

	input := fmt.Sprintf("protocol=%s\nhost=%s\nusername=%s\npassword=%s\n\n",
		u.Scheme, u.Host, username, password)

	cmd := exec.Command("git", "credential", "approve")
	cmd.Stdin = strings.NewReader(input)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git credential approve: %w", err)
	}
	return nil
}

// Erase deletes credentials from the keyring.
func Erase(serverURL, username string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("failed to parse server URL: %w", err)
	}

	input := fmt.Sprintf("protocol=%s\nhost=%s\n", u.Scheme, u.Host)
	if username != "" {
		input += fmt.Sprintf("username=%s\n", username)
	}
	input += "\n"

	cmd := exec.Command("git", "credential", "reject")
	cmd.Stdin = strings.NewReader(input)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run git credential reject: %w", err)
	}
	return nil
}
