package repo

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type RepoInfo struct {
	Owner string
	Name  string
	Host  string
}

// ParseRepoFlag parses "owner/name" format from --repo flag.
func ParseRepoFlag(flag string) (RepoInfo, error) {
	if flag == "" {
		return RepoInfo{}, fmt.Errorf("empty repo flag")
	}
	parts := strings.Split(flag, "/")
	if len(parts) != 2 {
		return RepoInfo{}, fmt.Errorf("invalid repo format %q: expected owner/name", flag)
	}
	if parts[0] == "" || parts[1] == "" {
		return RepoInfo{}, fmt.Errorf("invalid repo format %q: owner and name must not be empty", flag)
	}
	return RepoInfo{
		Owner: parts[0],
		Name:  parts[1],
		Host:  "github.com",
	}, nil
}

// ParseGitRemoteURL extracts owner/name from a git remote URL.
// Supports HTTPS (https://github.com/owner/repo.git) and SSH (git@github.com:owner/repo.git).
func ParseGitRemoteURL(remoteURL string) (RepoInfo, error) {
	if remoteURL == "" {
		return RepoInfo{}, fmt.Errorf("empty remote URL")
	}

	var host, path string

	// SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		trimmed := strings.TrimPrefix(remoteURL, "git@")
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx < 0 {
			return RepoInfo{}, fmt.Errorf("invalid SSH remote URL: %q", remoteURL)
		}
		host = trimmed[:colonIdx]
		path = trimmed[colonIdx+1:]
	} else if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return RepoInfo{}, fmt.Errorf("invalid remote URL %q: %w", remoteURL, err)
		}
		host = u.Host
		path = strings.TrimPrefix(u.Path, "/")
	} else {
		return RepoInfo{}, fmt.Errorf("unsupported remote URL format: %q", remoteURL)
	}

	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return RepoInfo{}, fmt.Errorf("could not extract owner/name from URL: %q", remoteURL)
	}

	return RepoInfo{
		Owner: parts[0],
		Name:  parts[1],
		Host:  host,
	}, nil
}

// DetectFromGitRemote detects the repo from the git remote "origin" in the current directory.
func DetectFromGitRemote() (RepoInfo, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return RepoInfo{}, fmt.Errorf("failed to detect git remote: %w (are you inside a git repository?)", err)
	}
	remoteURL := strings.TrimSpace(string(out))
	return ParseGitRemoteURL(remoteURL)
}

// Resolve resolves the repo info, preferring the --repo flag over git remote detection.
func Resolve(repoFlag string) (RepoInfo, error) {
	if repoFlag != "" {
		return ParseRepoFlag(repoFlag)
	}
	return DetectFromGitRemote()
}
