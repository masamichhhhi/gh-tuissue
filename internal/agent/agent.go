// Package agent launches background Claude Code sessions for an issue.
//
// An AgentAction from the config binds a key to a Claude Code skill. Pressing
// the key on an issue runs `claude --bg ... "/<skill> <args>"` in the
// directory gh-tuissue was started from. The session starts working
// immediately and shows up in `claude agents`, where the user can attach to
// it later.
package agent

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/masamichhhhi/gh-tuissue/internal/config"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

const (
	// DefaultArgs is the argument template used when an action has none.
	DefaultArgs = "{{number}}"
	// maxNameLen caps the background session name so it fits in tab titles.
	maxNameLen = 40
	// Binary is the Claude Code executable.
	Binary = "claude"
)

// Launch describes a background session that was started successfully.
type Launch struct {
	ID   string
	Name string
}

// Validate checks the configured actions and returns the usable ones plus a
// warning for every action that was skipped.
func Validate(actions []config.AgentAction) ([]config.AgentAction, []string) {
	var valid []config.AgentAction
	var warnings []string
	seen := make(map[string]bool)
	for _, a := range actions {
		switch {
		case len([]rune(a.Key)) != 1:
			warnings = append(warnings, fmt.Sprintf("agent %q: key must be a single character, got %q", a.Skill, a.Key))
		case strings.TrimSpace(a.Skill) == "":
			warnings = append(warnings, fmt.Sprintf("agent key %q: skill is empty", a.Key))
		case seen[a.Key]:
			warnings = append(warnings, fmt.Sprintf("agent key %q: duplicate key, %q skipped", a.Key, a.Skill))
		default:
			seen[a.Key] = true
			valid = append(valid, a)
		}
	}
	return valid, warnings
}

// Command returns the argv that launches a background Claude Code session
// running the action's skill against the issue.
func Command(action config.AgentAction, issue domain.Issue) []string {
	args := action.Args
	if args == "" {
		args = DefaultArgs
	}
	prompt := strings.TrimSpace("/" + strings.TrimPrefix(action.Skill, "/") + " " + Expand(args, issue))
	return []string{
		Binary,
		"--bg",
		"--permission-mode", "auto",
		"-n", SessionName(issue),
		prompt,
	}
}

// Expand replaces {{number}}, {{title}} and {{url}} in tmpl with issue data.
func Expand(tmpl string, issue domain.Issue) string {
	r := strings.NewReplacer(
		"{{number}}", fmt.Sprintf("%d", issue.Number),
		"{{title}}", issue.Title,
		"{{url}}", issue.URL,
	)
	return r.Replace(tmpl)
}

// SessionName builds "<number>-<title-slug>" truncated to maxNameLen.
func SessionName(issue domain.Issue) string {
	name := fmt.Sprintf("%d", issue.Number)
	if slug := Slugify(issue.Title); slug != "" {
		name += "-" + slug
	}
	if len(name) > maxNameLen {
		name = strings.TrimRight(name[:maxNameLen], "-")
	}
	return name
}

// Slugify lowercases s and replaces runs of non-alphanumerics with "-".
func Slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if r > unicode.MaxASCII {
				// Keep names ASCII so they are safe in shells and tab titles.
				continue
			}
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// Run executes argv in dir (the current directory when empty) and parses the
// id of the background session.
func Run(dir string, argv []string) (Launch, error) {
	if len(argv) == 0 {
		return Launch{}, errors.New("empty command")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Launch{}, fmt.Errorf("%s not found in PATH", argv[0])
		}
		msg := firstLine(string(out))
		if msg == "" {
			return Launch{}, err
		}
		return Launch{}, fmt.Errorf("%w: %s", err, msg)
	}
	return ParseOutput(string(out))
}

var (
	ansiRe         = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	backgroundedRe = regexp.MustCompile(`backgrounded\s*·\s*(\S+)\s*·\s*(.*)`)
)

// ParseOutput extracts the session id and name from `claude --bg` output.
func ParseOutput(out string) (Launch, error) {
	clean := ansiRe.ReplaceAllString(out, "")
	m := backgroundedRe.FindStringSubmatch(clean)
	if m == nil {
		return Launch{}, fmt.Errorf("unexpected claude output: %s", firstLine(clean))
	}
	return Launch{ID: m[1], Name: strings.TrimSpace(m[2])}, nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(ansiRe.ReplaceAllString(s, ""))
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return s
}
