package ui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestAppModel_InitialState(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	if app.currentView != ViewBoard {
		t.Errorf("initial view = %d, want ViewBoard (%d)", app.currentView, ViewBoard)
	}
}

func TestAppModel_QuitKey(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	msg := tea.KeyPressMsg{Code: 'q'}
	_, cmd := app.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
}

func TestAppModel_HelpToggle(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	msg := tea.KeyPressMsg{Code: '?'}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	if appModel.currentView != ViewHelp {
		t.Errorf("after ? key, view = %d, want ViewHelp (%d)", appModel.currentView, ViewHelp)
	}
}

func TestAppModel_EscFromHelp(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	app.currentView = ViewHelp
	app.prevView = ViewBoard
	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	if appModel.currentView != ViewBoard {
		t.Errorf("after Esc from help, view = %d, want ViewBoard (%d)", appModel.currentView, ViewBoard)
	}
}

func TestAppModel_WindowResize(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	if appModel.width != 120 || appModel.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", appModel.width, appModel.height)
	}
}

func TestAppModel_EditorResultNoService(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	app.currentView = ViewDetail
	app.lastEditType = editNone

	// Simulate editor result - issueSvc is nil so it returns early
	msg := editorResultMsg{content: "# Title\n\nDescription here"}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	// With nil issueSvc, handleNewIssue returns without doing anything
	if appModel.currentView != ViewDetail {
		t.Errorf("expected view to remain, got %d", appModel.currentView)
	}
}

func TestAppModel_EditorErrorReturnsToBoard(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	app.currentView = ViewDetail

	msg := editorResultMsg{err: fmt.Errorf("editor failed")}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	// On editor error, should return to board
	if appModel.currentView != ViewBoard {
		t.Errorf("expected ViewBoard after editor error, got %d", appModel.currentView)
	}
}

func TestAppModel_ProjectDataMsg(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", false)
	app.board.loading = true

	msg := projectDataMsg{
		info: sampleProjectInfo(),
		items: sampleProjectItems(),
	}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	if appModel.board.loading {
		t.Error("board should not be loading after projectDataMsg")
	}
	if len(appModel.board.columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(appModel.board.columns))
	}
}
