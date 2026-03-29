package ui

import (
	"os"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

// launchEditor opens $EDITOR with a temp file and returns the content.
func launchEditor(initialContent string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		tmpFile, err := os.CreateTemp("", "gh-tuissue-*.md")
		if err != nil {
			return editorResultMsg{err: err}
		}
		tmpPath := tmpFile.Name()

		if initialContent != "" {
			tmpFile.WriteString(initialContent)
		}
		tmpFile.Close()

		cmd := exec.Command(editor, tmpPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			os.Remove(tmpPath)
			return editorResultMsg{err: err}
		}

		content, err := os.ReadFile(tmpPath)
		os.Remove(tmpPath)
		if err != nil {
			return editorResultMsg{err: err}
		}

		return editorResultMsg{content: string(content)}
	}
}

type editorResultMsg struct {
	content string
	err     error
}
