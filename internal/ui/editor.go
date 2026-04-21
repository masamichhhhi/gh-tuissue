package ui

import (
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

// launchEditor opens $EDITOR with a temp file and returns the content.
// It uses tea.ExecProcess to properly pause Bubble Tea's input handling
// while the editor is running, avoiding stdin read conflicts.
func launchEditor(initialContent string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "gh-tuissue-*.md")
	if err != nil {
		return func() tea.Msg {
			return editorResultMsg{err: err}
		}
	}
	tmpPath := tmpFile.Name()

	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			return func() tea.Msg {
				return editorResultMsg{err: err}
			}
		}
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return func() tea.Msg {
			return editorResultMsg{err: err}
		}
	}

	c := exec.Command(editor, tmpPath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			_ = os.Remove(tmpPath)
			return editorResultMsg{err: err}
		}

		content, readErr := os.ReadFile(tmpPath)
		_ = os.Remove(tmpPath)
		if readErr != nil {
			return editorResultMsg{err: readErr}
		}

		return editorResultMsg{content: string(content)}
	})
}

type editorResultMsg struct {
	content string
	err     error
}
