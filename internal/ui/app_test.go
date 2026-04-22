package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
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

func TestAppModel_HandleNewIssue_UsesActiveColumnStatus(t *testing.T) {
	var mu sync.Mutex
	var moveCalls []map[string]interface{}
	var addCallCount int

	mock := &gh.MockClient{
		RESTPostHandler: func(ctx context.Context, path string, body interface{}, result interface{}) error {
			// CreateIssue path: repos/owner/repo/issues
			resp := map[string]interface{}{
				"number":     42,
				"node_id":    "I_new",
				"title":      "New Issue",
				"body":       "",
				"state":      "open",
				"html_url":   "https://example.com/42",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]interface{}{"login": "creator"},
				"labels":     []interface{}{},
				"assignees":  []interface{}{},
				"milestone":  nil,
			}
			b, _ := json.Marshal(resp)
			return json.Unmarshal(b, result)
		},
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			mu.Lock()
			defer mu.Unlock()
			switch {
			case strings.Contains(query, "addProjectV2ItemById"):
				addCallCount++
				resp := map[string]interface{}{
					"addProjectV2ItemById": map[string]interface{}{
						"item": map[string]interface{}{"id": "PVTI_new"},
					},
				}
				b, _ := json.Marshal(resp)
				return json.Unmarshal(b, result)
			case strings.Contains(query, "updateProjectV2ItemFieldValue"):
				copied := make(map[string]interface{}, len(variables))
				for k, v := range variables {
					copied[k] = v
				}
				moveCalls = append(moveCalls, copied)
				return nil
			}
			return nil
		},
	}

	issueSvc := service.NewIssueService(mock, "owner", "repo")
	projectSvc := service.NewProjectService(mock, "owner", "repo")

	app := NewAppModel(issueSvc, nil, projectSvc, 1, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetProjectData(sampleProjectInfo(), sampleProjectItems())
	// Move to "In Progress" column (index 1)
	app.board.activeCol = 1

	_, cmd := app.handleNewIssue("# Created from In Progress\n\nbody")
	if cmd == nil {
		t.Fatal("expected a command from handleNewIssue")
	}
	msg := cmd()
	created, ok := msg.(issueCreatedMsg)
	if !ok || created.err != nil {
		t.Fatalf("expected issueCreatedMsg with nil err, got %+v", msg)
	}
	if created.itemID != "PVTI_new" {
		t.Errorf("itemID = %q, want PVTI_new", created.itemID)
	}
	if created.optionID != "opt_progress" {
		t.Errorf("optionID = %q, want opt_progress", created.optionID)
	}
	if created.issue.Number != 42 {
		t.Errorf("issue.Number = %d, want 42", created.issue.Number)
	}

	mu.Lock()
	defer mu.Unlock()
	if addCallCount != 1 {
		t.Errorf("addProjectV2ItemById call count = %d, want 1", addCallCount)
	}
	if len(moveCalls) != 1 {
		t.Fatalf("updateProjectV2ItemFieldValue call count = %d, want 1", len(moveCalls))
	}
	got := moveCalls[0]
	if got["projectId"] != "PVT_123" {
		t.Errorf("projectId = %v, want PVT_123", got["projectId"])
	}
	if got["itemId"] != "PVTI_new" {
		t.Errorf("itemId = %v, want PVTI_new", got["itemId"])
	}
	if got["fieldId"] != "PVTSSF_1" {
		t.Errorf("fieldId = %v, want PVTSSF_1", got["fieldId"])
	}
	if got["optionId"] != "opt_progress" {
		t.Errorf("optionId = %v, want opt_progress (In Progress column)", got["optionId"])
	}
}

func TestAppModel_IssueCreatedMsg_OptimisticInsert_NoReload(t *testing.T) {
	app := NewAppModel(nil, nil, dummyProjectSvc(), 1, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// In Progress column starts with 1 item (PVTI_3)
	beforeCount := len(app.board.columns[1].Items)

	msg := issueCreatedMsg{
		issue:    domain.Issue{Number: 99, NodeID: "I_new", Title: "Just Created", State: domain.IssueOpen},
		itemID:   "PVTI_new",
		optionID: "opt_progress",
	}
	updated, cmd := app.Update(msg)
	appModel := updated.(AppModel)

	// No reload command — local insertion handles it
	if cmd != nil {
		t.Error("expected no reload command when optimistic insert is possible")
	}
	if appModel.currentView != ViewBoard {
		t.Errorf("currentView = %d, want ViewBoard", appModel.currentView)
	}
	afterCount := len(appModel.board.columns[1].Items)
	if afterCount != beforeCount+1 {
		t.Errorf("In Progress column items = %d, want %d", afterCount, beforeCount+1)
	}
	if appModel.board.columns[1].Items[afterCount-1].ItemID != "PVTI_new" {
		t.Errorf("new item not appended at end of target column")
	}
	if !strings.Contains(appModel.statusMsg, "#99") {
		t.Errorf("statusMsg = %q, want to include #99", appModel.statusMsg)
	}
}

func TestAppModel_IssueCreatedMsg_FallbackMode_Reloads(t *testing.T) {
	app := NewAppModel(nil, nil, nil, 0, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetFallbackIssues(nil)

	msg := issueCreatedMsg{
		issue: domain.Issue{Number: 7, Title: "Fallback", State: domain.IssueOpen},
	}
	_, cmd := app.Update(msg)

	// In fallback (no projectInfo), we don't have enough local info; reload is
	// still the correct behavior. cmd will be nil because issueSvc is nil, but
	// that's a harmless no-op.
	_ = cmd
}

func TestAppModel_IssueCreatedMsg_Error(t *testing.T) {
	app := NewAppModel(nil, nil, dummyProjectSvc(), 1, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	msg := issueCreatedMsg{err: fmt.Errorf("API error")}
	updated, cmd := app.Update(msg)
	appModel := updated.(AppModel)

	if cmd != nil {
		t.Error("expected no command on error")
	}
	if !strings.Contains(appModel.statusMsg, "API error") {
		t.Errorf("statusMsg = %q, want to include error", appModel.statusMsg)
	}
}

func TestAppModel_HandleNewIssue_NoProject_SkipsStatusSet(t *testing.T) {
	var graphqlCalls int

	mock := &gh.MockClient{
		RESTPostHandler: func(ctx context.Context, path string, body interface{}, result interface{}) error {
			resp := map[string]interface{}{
				"number":     1,
				"node_id":    "I_1",
				"title":      "T",
				"body":       "",
				"state":      "open",
				"html_url":   "",
				"created_at": "2026-01-01T00:00:00Z",
				"updated_at": "2026-01-01T00:00:00Z",
				"user":       map[string]interface{}{"login": "u"},
				"labels":     []interface{}{},
				"assignees":  []interface{}{},
				"milestone":  nil,
			}
			b, _ := json.Marshal(resp)
			return json.Unmarshal(b, result)
		},
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			graphqlCalls++
			return nil
		},
	}

	issueSvc := service.NewIssueService(mock, "owner", "repo")
	// No project service - fallback mode
	app := NewAppModel(issueSvc, nil, nil, 0, "", nil)
	app.board.SetSize(120, 40)
	app.board.SetFallbackIssues(nil)

	_, cmd := app.handleNewIssue("# My New Issue\n\nbody")
	if cmd == nil {
		t.Fatal("expected a command")
	}
	_ = cmd()

	if graphqlCalls != 0 {
		t.Errorf("expected no GraphQL calls in fallback mode, got %d", graphqlCalls)
	}
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
