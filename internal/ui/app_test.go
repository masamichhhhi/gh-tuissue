package ui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	"github.com/masamichhhhi/gh-tuissue/internal/service"
)

func TestAppModel_InitialState(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	if app.currentView != ViewBoard {
		t.Errorf("initial view = %d, want ViewBoard (%d)", app.currentView, ViewBoard)
	}
}

func TestAppModel_QKeyDoesNotQuit(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	msg := tea.KeyPressMsg{Code: 'q'}
	_, cmd := app.Update(msg)
	if cmd != nil {
		t.Fatal("expected no quit command from q key, but got one")
	}
}

func TestAppModel_HelpToggle(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	msg := tea.KeyPressMsg{Code: '?'}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	if appModel.currentView != ViewHelp {
		t.Errorf("after ? key, view = %d, want ViewHelp (%d)", appModel.currentView, ViewHelp)
	}
}

func TestAppModel_EscFromHelp(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
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
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	if appModel.width != 120 || appModel.height != 40 {
		t.Errorf("size = %dx%d, want 120x40", appModel.width, appModel.height)
	}
}

func TestAppModel_EditorResultNoService(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
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
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	app.currentView = ViewDetail

	msg := editorResultMsg{err: fmt.Errorf("editor failed")}
	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)
	// On editor error, should return to board
	if appModel.currentView != ViewBoard {
		t.Errorf("expected ViewBoard after editor error, got %d", appModel.currentView)
	}
}

func dummyProjectSvc() *service.ProjectService {
	return service.NewProjectService(nil, "owner", "repo")
}

func TestAppModel_OptimisticStatusMove(t *testing.T) {
	app := NewAppModel(nil, nil, dummyProjectSvc(), 1, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Simulate board requesting status move right (Shift+L)
	app.board.wantStatusMove = 1

	// handleStatusMove should optimistically move the item before API call
	updated, cmd := app.handleStatusMove()
	appModel := updated.(AppModel)

	// The item should already be in the target column (optimistic update)
	if len(appModel.board.columns[0].Items) != 1 {
		t.Errorf("Todo column has %d items, want 1 (item should have moved)", len(appModel.board.columns[0].Items))
	}
	if len(appModel.board.columns[1].Items) != 2 {
		t.Errorf("In Progress column has %d items, want 2 (item should have moved here)", len(appModel.board.columns[1].Items))
	}

	// A command should be returned for the async API call
	if cmd == nil {
		t.Fatal("expected a command for async API call")
	}
}

func TestAppModel_StatusMoveMsg_Success(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Simulate a successful statusMoveMsg (no error)
	msg := statusMoveMsg{
		err:            nil,
		itemID:         "PVTI_1",
		originalStatus: "opt_todo",
		originalColIdx: 0,
	}

	updated, cmd := app.Update(msg)
	appModel := updated.(AppModel)

	// On success, no reload should happen (cmd should be nil)
	if cmd != nil {
		t.Error("expected no command on success (no reload)")
	}
	if appModel.statusMsg != "Status updated" {
		t.Errorf("statusMsg = %q, want %q", appModel.statusMsg, "Status updated")
	}
}

func TestAppModel_StatusMoveMsg_Failure_Rollback(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// First, optimistically move PVTI_1 from Todo (0) to In Progress (1)
	app.board.MoveItemToColumn("PVTI_1", 0, 1)

	// Verify optimistic state
	if len(app.board.columns[0].Items) != 1 {
		t.Fatalf("expected 1 item in Todo after optimistic move, got %d", len(app.board.columns[0].Items))
	}

	// Simulate a failed statusMoveMsg
	msg := statusMoveMsg{
		err:            fmt.Errorf("API error"),
		itemID:         "PVTI_1",
		originalStatus: "opt_todo",
		originalColIdx: 0,
	}

	updated, _ := app.Update(msg)
	appModel := updated.(AppModel)

	// Item should be rolled back to original column
	if len(appModel.board.columns[0].Items) != 2 {
		t.Errorf("Todo column has %d items after rollback, want 2", len(appModel.board.columns[0].Items))
	}
	if len(appModel.board.columns[1].Items) != 1 {
		t.Errorf("In Progress column has %d items after rollback, want 1", len(appModel.board.columns[1].Items))
	}
	if appModel.statusMsg != "Status move failed: API error" {
		t.Errorf("statusMsg = %q, want error message", appModel.statusMsg)
	}
}

func TestAppModel_StatusMoveNoProject(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", nil)
	app.board.SetSize(120, 40)
	app.board.wantStatusMove = 1

	updated, _ := app.handleStatusMove()
	appModel := updated.(AppModel)

	if appModel.statusMsg != "Status move not available (no project)" {
		t.Errorf("statusMsg = %q, want no project message", appModel.statusMsg)
	}
}

func ptrProjectInfo(info domain.ProjectInfo) *domain.ProjectInfo {
	return &info
}

func TestAppModel_ProjectDataMsg(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 1, "", nil)
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
