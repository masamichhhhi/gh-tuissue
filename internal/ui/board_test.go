package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

func sampleProjectInfo() domain.ProjectInfo {
	return domain.ProjectInfo{
		ID:    "PVT_123",
		Title: "Test Project",
		StatusField: domain.StatusField{
			ID:   "PVTSSF_1",
			Name: "Status",
			Options: []domain.StatusOption{
				{ID: "opt_todo", Name: "Todo"},
				{ID: "opt_progress", Name: "In Progress"},
				{ID: "opt_done", Name: "Done"},
			},
		},
	}
}

func sampleProjectItems() []domain.ProjectItem {
	return []domain.ProjectItem{
		{ItemID: "PVTI_1", Issue: domain.Issue{Number: 1, Title: "First Issue", State: domain.IssueOpen, Labels: []domain.Label{{Name: "bug", Color: "d73a4a"}}}, StatusID: "opt_todo"},
		{ItemID: "PVTI_2", Issue: domain.Issue{Number: 2, Title: "Second Issue", State: domain.IssueOpen, Assignees: []domain.User{{Login: "dev1"}}}, StatusID: "opt_todo"},
		{ItemID: "PVTI_3", Issue: domain.Issue{Number: 3, Title: "In Progress Issue", State: domain.IssueOpen}, StatusID: "opt_progress"},
		{ItemID: "PVTI_4", Issue: domain.Issue{Number: 4, Title: "Done Issue", State: domain.IssueClosed}, StatusID: "opt_done"},
	}
}

func TestBoardModel_SetProjectData(t *testing.T) {
	board := NewBoardModel()
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	if len(board.columns) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(board.columns))
	}
	if board.columns[0].Name != "Todo" {
		t.Errorf("column[0].Name = %q, want %q", board.columns[0].Name, "Todo")
	}
	if len(board.columns[0].Items) != 2 {
		t.Errorf("Todo column has %d items, want 2", len(board.columns[0].Items))
	}
	if len(board.columns[1].Items) != 1 {
		t.Errorf("In Progress column has %d items, want 1", len(board.columns[1].Items))
	}
	if len(board.columns[2].Items) != 1 {
		t.Errorf("Done column has %d items, want 1", len(board.columns[2].Items))
	}
}

func TestBoardModel_SetFallbackIssues(t *testing.T) {
	board := NewBoardModel()
	issues := []domain.Issue{
		{Number: 1, Title: "Open", State: domain.IssueOpen},
		{Number: 2, Title: "Closed", State: domain.IssueClosed},
	}
	board.SetFallbackIssues(issues)

	if len(board.columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(board.columns))
	}
	if board.columns[0].Name != "Open" {
		t.Errorf("column[0].Name = %q, want %q", board.columns[0].Name, "Open")
	}
	if len(board.columns[0].Items) != 1 {
		t.Errorf("Open column has %d items, want 1", len(board.columns[0].Items))
	}
	if len(board.columns[1].Items) != 1 {
		t.Errorf("Closed column has %d items, want 1", len(board.columns[1].Items))
	}
}

func TestBoardModel_Navigation_VimKeys(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Start in first column (Todo)
	if board.activeCol != 0 {
		t.Errorf("expected activeCol=0, got %d", board.activeCol)
	}

	// Move down with j
	board, _ = board.Update(tea.KeyPressMsg{Code: 'j'})
	if board.cursorIndex[0] != 1 {
		t.Errorf("cursor = %d, want 1", board.cursorIndex[0])
	}

	// Move up with k
	board, _ = board.Update(tea.KeyPressMsg{Code: 'k'})
	if board.cursorIndex[0] != 0 {
		t.Errorf("cursor = %d, want 0", board.cursorIndex[0])
	}

	// Move right with l
	board, _ = board.Update(tea.KeyPressMsg{Code: 'l'})
	if board.activeCol != 1 {
		t.Errorf("expected activeCol=1, got %d", board.activeCol)
	}

	// Move right again
	board, _ = board.Update(tea.KeyPressMsg{Code: 'l'})
	if board.activeCol != 2 {
		t.Errorf("expected activeCol=2, got %d", board.activeCol)
	}

	// Move right at rightmost - should stay
	board, _ = board.Update(tea.KeyPressMsg{Code: 'l'})
	if board.activeCol != 2 {
		t.Errorf("expected activeCol=2 (should not go beyond), got %d", board.activeCol)
	}

	// Move left with h
	board, _ = board.Update(tea.KeyPressMsg{Code: 'h'})
	if board.activeCol != 1 {
		t.Errorf("expected activeCol=1, got %d", board.activeCol)
	}
}

func TestBoardModel_Navigation_ArrowKeys(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Move down with arrow
	board, _ = board.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if board.cursorIndex[0] != 1 {
		t.Errorf("cursor = %d, want 1 after arrow down", board.cursorIndex[0])
	}

	// Move up with arrow
	board, _ = board.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if board.cursorIndex[0] != 0 {
		t.Errorf("cursor = %d, want 0 after arrow up", board.cursorIndex[0])
	}

	// Move right with arrow
	board, _ = board.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if board.activeCol != 1 {
		t.Errorf("expected activeCol=1 after arrow right, got %d", board.activeCol)
	}

	// Move left with arrow
	board, _ = board.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if board.activeCol != 0 {
		t.Errorf("expected activeCol=0 after arrow left, got %d", board.activeCol)
	}

	// Left at leftmost - should stay
	board, _ = board.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if board.activeCol != 0 {
		t.Errorf("expected activeCol=0 (should not go below), got %d", board.activeCol)
	}
}

func TestBoardModel_SelectIssue(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	board, _ = board.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if board.selectedIssue == nil {
		t.Fatal("expected selectedIssue to be set")
	}
	if board.selectedIssue.Number != 1 {
		t.Errorf("selected issue = #%d, want #1", board.selectedIssue.Number)
	}
}

func TestBoardModel_StatusMove(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Shift+L should request status move right
	board, _ = board.Update(tea.KeyPressMsg{Code: 'L'})
	if board.wantStatusMove != 1 {
		t.Errorf("wantStatusMove = %d, want 1", board.wantStatusMove)
	}

	board.wantStatusMove = 0

	// Shift+H should request status move left
	board, _ = board.Update(tea.KeyPressMsg{Code: 'H'})
	if board.wantStatusMove != -1 {
		t.Errorf("wantStatusMove = %d, want -1", board.wantStatusMove)
	}
}

func TestBoardModel_Filter(t *testing.T) {
	board := NewBoardModel()
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	board.SetFilter(FilterState{Labels: []string{"bug"}})
	// Only item 1 has "bug" label, it's in Todo
	total := 0
	for _, col := range board.columns {
		total += len(col.Items)
	}
	if total != 1 {
		t.Errorf("expected 1 filtered item total, got %d", total)
	}
}

func TestBoardModel_EmptyFilter(t *testing.T) {
	board := NewBoardModel()
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())
	board.SetFilter(FilterState{Labels: []string{"nonexistent"}})
	total := 0
	for _, col := range board.columns {
		total += len(col.Items)
	}
	if total != 0 {
		t.Errorf("expected 0 filtered items, got %d", total)
	}
}

func TestBoardModel_View_Loading(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(80, 24)
	view := board.View()
	if view == "" {
		t.Error("expected non-empty view for loading state")
	}
}

func TestBoardModel_CursorBounds(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Try moving up from 0 - should stay at 0
	board, _ = board.Update(tea.KeyPressMsg{Code: 'k'})
	if board.cursorIndex[0] != 0 {
		t.Errorf("cursor should not go below 0, got %d", board.cursorIndex[0])
	}

	// Move to end (Todo has 2 items)
	board, _ = board.Update(tea.KeyPressMsg{Code: 'j'})
	board, _ = board.Update(tea.KeyPressMsg{Code: 'j'}) // beyond end
	if board.cursorIndex[0] != 1 { // max index = 1
		t.Errorf("cursor = %d, want 1 (max)", board.cursorIndex[0])
	}
}

func TestBoardModel_SelectedItem(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	item := board.SelectedItem()
	if item == nil {
		t.Fatal("expected SelectedItem to return item")
	}
	if item.ItemID != "PVTI_1" {
		t.Errorf("SelectedItem().ItemID = %q, want %q", item.ItemID, "PVTI_1")
	}
}
