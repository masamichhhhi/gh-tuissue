package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// resolveToken retrieves the GitHub authentication token.
// Priority: GH_TOKEN env > GITHUB_TOKEN env > gh auth token command.
func resolveToken() (string, error) {
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token, nil
	}
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get token from gh CLI: %w", err)
	}
	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf("gh auth token returned empty result")
	}
	return token, nil
}
