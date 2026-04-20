package ui

import (
	"errors"
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
	cleanupWithError := func(baseErr error) tea.Msg {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			baseErr = errors.Join(baseErr, removeErr)
		}
		return editorResultMsg{err: baseErr}
	}

	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			if closeErr := tmpFile.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
			return func() tea.Msg {
				return cleanupWithError(err)
			}
		}
	}
	if err := tmpFile.Close(); err != nil {
		return func() tea.Msg {
			return cleanupWithError(err)
		}
	}

	c := exec.Command(editor, tmpPath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			return cleanupWithError(err)
		}

		content, readErr := os.ReadFile(tmpPath)
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			if readErr != nil {
				return editorResultMsg{err: errors.Join(readErr, removeErr)}
			}
			return editorResultMsg{err: removeErr}
		}
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
