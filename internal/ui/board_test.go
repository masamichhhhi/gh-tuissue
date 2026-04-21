package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-runewidth"
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
	board, _ = board.Update(tea.KeyPressMsg{Code: 'l', Mod: tea.ModShift})
	if board.wantStatusMove != 1 {
		t.Errorf("wantStatusMove = %d, want 1", board.wantStatusMove)
	}

	board.wantStatusMove = 0

	// Shift+H should request status move left
	board, _ = board.Update(tea.KeyPressMsg{Code: 'h', Mod: tea.ModShift})
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

func TestBoardModel_HideColumn(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// 3 columns: Todo, In Progress, Done. Active = 0 (Todo)
	if board.HiddenCount() != 0 {
		t.Errorf("HiddenCount = %d, want 0", board.HiddenCount())
	}

	// Hide current column (Todo)
	board, _ = board.Update(tea.KeyPressMsg{Code: 'd'})
	if board.HiddenCount() != 1 {
		t.Errorf("HiddenCount = %d, want 1 after hide", board.HiddenCount())
	}
	// Active column should move to next visible (In Progress = 1)
	if board.activeCol != 1 {
		t.Errorf("activeCol = %d, want 1 after hiding col 0", board.activeCol)
	}
}

func TestBoardModel_HideLastColumnRefused(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Hide first two columns
	board, _ = board.Update(tea.KeyPressMsg{Code: 'd'}) // hide Todo
	board, _ = board.Update(tea.KeyPressMsg{Code: 'd'}) // hide In Progress

	// Try to hide the last one (Done)
	board, _ = board.Update(tea.KeyPressMsg{Code: 'd'})
	if board.HiddenCount() != 2 {
		t.Errorf("HiddenCount = %d, want 2 (should not allow hiding last)", board.HiddenCount())
	}
	if board.wantStatusMsg == "" {
		t.Error("expected a status message when hiding last column is refused")
	}
}

func TestBoardModel_RestoreAllColumns(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	board, _ = board.Update(tea.KeyPressMsg{Code: 'd'}) // hide one
	board, _ = board.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModShift}) // restore all

	if board.HiddenCount() != 0 {
		t.Errorf("HiddenCount = %d, want 0 after restore", board.HiddenCount())
	}
}

func TestBoardModel_MoveItemToColumn(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Initial state: Todo has 2 items (PVTI_1, PVTI_2), In Progress has 1 (PVTI_3)
	if len(board.columns[0].Items) != 2 {
		t.Fatalf("Todo column has %d items, want 2", len(board.columns[0].Items))
	}
	if len(board.columns[1].Items) != 1 {
		t.Fatalf("In Progress column has %d items, want 1", len(board.columns[1].Items))
	}

	// Move PVTI_1 from Todo (col 0) to In Progress (col 1)
	ok := board.MoveItemToColumn("PVTI_1", 0, 1)
	if !ok {
		t.Fatal("MoveItemToColumn returned false")
	}

	// Todo should now have 1 item, In Progress should have 2
	if len(board.columns[0].Items) != 1 {
		t.Errorf("Todo column has %d items after move, want 1", len(board.columns[0].Items))
	}
	if len(board.columns[1].Items) != 2 {
		t.Errorf("In Progress column has %d items after move, want 2", len(board.columns[1].Items))
	}

	// The moved item should have its StatusID updated to the target column's OptionID
	found := false
	for _, item := range board.columns[1].Items {
		if item.ItemID == "PVTI_1" {
			found = true
			if item.StatusID != "opt_progress" {
				t.Errorf("moved item StatusID = %q, want %q", item.StatusID, "opt_progress")
			}
		}
	}
	if !found {
		t.Error("moved item PVTI_1 not found in target column")
	}

	// allItems should also be updated
	for _, item := range board.allItems {
		if item.ItemID == "PVTI_1" {
			if item.StatusID != "opt_progress" {
				t.Errorf("allItems StatusID = %q, want %q", item.StatusID, "opt_progress")
			}
		}
	}
}

func TestBoardModel_MoveItemToColumn_CursorFollows(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// activeCol = 0 (Todo), cursor at 0 (PVTI_1)
	board.activeCol = 0
	board.cursorIndex[0] = 0

	board.MoveItemToColumn("PVTI_1", 0, 1)

	// Cursor should follow to target column
	if board.activeCol != 1 {
		t.Errorf("activeCol = %d, want 1 (should follow moved item)", board.activeCol)
	}
	// Cursor should point to the moved item (appended at end of target column)
	targetItems := board.columns[1].Items
	expectedIdx := len(targetItems) - 1
	if board.cursorIndex[1] != expectedIdx {
		t.Errorf("cursorIndex[1] = %d, want %d", board.cursorIndex[1], expectedIdx)
	}
}

func TestBoardModel_MoveItemToColumn_InvalidItem(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Try to move non-existent item
	ok := board.MoveItemToColumn("NONEXISTENT", 0, 1)
	if ok {
		t.Error("MoveItemToColumn should return false for non-existent item")
	}
}

func TestBoardModel_RollbackItemMove(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Move PVTI_1 from Todo to In Progress
	board.MoveItemToColumn("PVTI_1", 0, 1)

	// Verify it moved
	if len(board.columns[0].Items) != 1 {
		t.Fatalf("expected 1 item in Todo after move, got %d", len(board.columns[0].Items))
	}

	// Rollback: move it back to Todo with original StatusID
	ok := board.RollbackItemMove("PVTI_1", "opt_todo", 0)
	if !ok {
		t.Fatal("RollbackItemMove returned false")
	}

	// Todo should have 2 items again, In Progress should have 1
	if len(board.columns[0].Items) != 2 {
		t.Errorf("Todo column has %d items after rollback, want 2", len(board.columns[0].Items))
	}
	if len(board.columns[1].Items) != 1 {
		t.Errorf("In Progress column has %d items after rollback, want 1", len(board.columns[1].Items))
	}

	// StatusID should be restored
	for _, item := range board.columns[0].Items {
		if item.ItemID == "PVTI_1" {
			if item.StatusID != "opt_todo" {
				t.Errorf("rolled back item StatusID = %q, want %q", item.StatusID, "opt_todo")
			}
		}
	}

	// allItems should also be restored
	for _, item := range board.allItems {
		if item.ItemID == "PVTI_1" {
			if item.StatusID != "opt_todo" {
				t.Errorf("allItems StatusID after rollback = %q, want %q", item.StatusID, "opt_todo")
			}
		}
	}
}

func TestBoardModel_MoveItemToColumn_BoundaryLeft(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Try to move from col 0 to col -1 (out of bounds)
	ok := board.MoveItemToColumn("PVTI_1", 0, -1)
	if ok {
		t.Error("MoveItemToColumn should return false for out-of-bounds target")
	}
}

func TestBoardModel_MoveItemToColumn_BoundaryRight(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Try to move from col 2 to col 3 (out of bounds)
	ok := board.MoveItemToColumn("PVTI_4", 2, 3)
	if ok {
		t.Error("MoveItemToColumn should return false for out-of-bounds target")
	}
}

func TestBoardModel_InsertItem_MatchingColumn(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	beforeAll := len(board.allItems)
	beforeInProgress := len(board.columns[1].Items)

	board.InsertItem("PVTI_new", "opt_progress", domain.Issue{Number: 99, Title: "New"})

	if len(board.allItems) != beforeAll+1 {
		t.Errorf("allItems len = %d, want %d", len(board.allItems), beforeAll+1)
	}
	if len(board.columns[1].Items) != beforeInProgress+1 {
		t.Errorf("In Progress items = %d, want %d", len(board.columns[1].Items), beforeInProgress+1)
	}
	last := board.columns[1].Items[len(board.columns[1].Items)-1]
	if last.ItemID != "PVTI_new" || last.StatusID != "opt_progress" {
		t.Errorf("last item = %+v, want itemID=PVTI_new, statusID=opt_progress", last)
	}
}

func TestBoardModel_InsertItem_UnknownOption_FallsBackToFirst(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	beforeFirst := len(board.columns[0].Items)
	board.InsertItem("PVTI_new", "opt_unknown", domain.Issue{Number: 99, Title: "New"})

	if len(board.columns[0].Items) != beforeFirst+1 {
		t.Errorf("first column items = %d, want %d (fallback)", len(board.columns[0].Items), beforeFirst+1)
	}
}

func TestBoardModel_InsertItem_FilteredOut(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())
	board.SetFilter(FilterState{Labels: []string{"bug"}})

	beforeAll := len(board.allItems)
	colCountsBefore := make([]int, len(board.columns))
	for i, c := range board.columns {
		colCountsBefore[i] = len(c.Items)
	}

	// New issue has no "bug" label, so it should be in allItems but no column.
	board.InsertItem("PVTI_new", "opt_progress", domain.Issue{Number: 99, Title: "No Label"})

	if len(board.allItems) != beforeAll+1 {
		t.Errorf("allItems should grow, got %d", len(board.allItems))
	}
	for i, c := range board.columns {
		if len(c.Items) != colCountsBefore[i] {
			t.Errorf("column %d items changed: %d -> %d (should be unchanged due to filter)", i, colCountsBefore[i], len(c.Items))
		}
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		wantFull bool // true if output should equal input (no truncation)
	}{
		{
			name:     "ASCII short enough",
			input:    "hello",
			maxWidth: 20,
			wantFull: true,
		},
		{
			name:     "ASCII exact fit",
			input:    "hello",
			maxWidth: 5,
			wantFull: true,
		},
		{
			name:     "ASCII truncated",
			input:    "hello world this is long",
			maxWidth: 10,
		},
		{
			name:     "Japanese short enough",
			input:    "バグ修正",
			maxWidth: 20,
			wantFull: true,
		},
		{
			name:     "Japanese truncated",
			input:    "日本語のタイトルが長い場合のテスト",
			maxWidth: 16,
		},
		{
			name:     "Mixed ASCII and Japanese",
			input:    "Fix: 日本語テキストの問題を修正",
			maxWidth: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateText(tt.input, tt.maxWidth)

			if tt.wantFull {
				if got != tt.input {
					t.Errorf("expected no truncation, got %q", got)
				}
				return
			}

			// Truncated result must end with "..."
			if !strings.HasSuffix(got, "...") {
				t.Errorf("truncated text should end with '...', got %q", got)
			}

			// Display width must not exceed maxWidth
			w := runewidth.StringWidth(got)
			if w > tt.maxWidth {
				t.Errorf("display width %d exceeds maxWidth %d, got %q", w, tt.maxWidth, got)
			}

			// Must not contain replacement character (the bug we fixed)
			if strings.Contains(got, "�") {
				t.Errorf("truncated text contains replacement character: %q", got)
			}
		})
	}
}

func TestBoardModel_NavigationSkipsHidden(t *testing.T) {
	board := NewBoardModel()
	board.SetSize(120, 40)
	board.SetProjectData(sampleProjectInfo(), sampleProjectItems())

	// Move to col 1 (In Progress)
	board, _ = board.Update(tea.KeyPressMsg{Code: 'l'})
	if board.activeCol != 1 {
		t.Fatalf("activeCol = %d, want 1", board.activeCol)
	}

	// Hide col 1 (In Progress) - should move to col 2 (Done)
	board, _ = board.Update(tea.KeyPressMsg{Code: 'd'})
	if board.activeCol != 2 {
		t.Errorf("activeCol = %d, want 2 after hiding col 1", board.activeCol)
	}

	// Navigate left should skip hidden col 1 and go to col 0
	board, _ = board.Update(tea.KeyPressMsg{Code: 'h'})
	if board.activeCol != 0 {
		t.Errorf("activeCol = %d, want 0 (skipping hidden col 1)", board.activeCol)
	}
}
