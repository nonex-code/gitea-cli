package main

import (
	"gitea-cli/cmd"
)

// Build-time variables, injected via -ldflags by GoReleaser.
// Example:
//
//	go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc123 -X main.date=2026-01-01"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	cmd.Execute()
}
