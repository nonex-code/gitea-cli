package credential

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test credential interaction logic by faking the git executable.
func TestGetFromCredentialHelper(t *testing.T) {
	dir := t.TempDir()

	// Fake git script that simulates the output of git credential fill
	script := `#!/bin/sh
cat > /dev/null  # read stdin
printf 'username=testuser\npassword=secrettoken\n'
`
	gitPath := filepath.Join(dir, "git")
	if err := os.WriteFile(gitPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	// Put the fake git at the front of PATH
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cred, err := Get("https://gitea.example.com", "testuser")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if cred.Username != "testuser" {
		t.Errorf("expected username=testuser, got %q", cred.Username)
	}
	if cred.Password != "secrettoken" {
		t.Errorf("expected password=secrettoken, got %q", cred.Password)
	}
	if cred.Host != "gitea.example.com" {
		t.Errorf("expected host=gitea.example.com, got %q", cred.Host)
	}
}

func TestGetInvalidURL(t *testing.T) {
	_, err := Get("://invalid", "")
	if err == nil {
		t.Error("expected an error when parsing an invalid URL")
	}
	if !strings.Contains(err.Error(), "failed to parse server URL") {
		t.Errorf("expected error message to contain URL parse hint, got: %v", err)
	}
}
