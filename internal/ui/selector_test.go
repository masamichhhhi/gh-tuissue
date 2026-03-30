package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSelectorModel_MultiSelect(t *testing.T) {
	items := []SelectorItem{
		{ID: "1", Name: "bug", Selected: true},
		{ID: "2", Name: "feature", Selected: false},
		{ID: "3", Name: "docs", Selected: false},
	}

	sel := NewSelectorModel("Labels", items, true)

	if !sel.active {
		t.Fatal("expected selector to be active")
	}
	if sel.cursor != 0 {
		t.Errorf("cursor = %d, want 0", sel.cursor)
	}

	// Move down
	sel, _ = sel.Update(tea.KeyPressMsg{Code: 'j'})
	if sel.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after j", sel.cursor)
	}

	// Toggle with space (multi-select)
	sel, _ = sel.Update(tea.KeyPressMsg{Code: ' '})
	if !sel.items[1].Selected {
		t.Error("expected item 1 to be selected after space toggle")
	}

	// Confirm
	sel, _ = sel.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !sel.confirmed {
		t.Error("expected confirmed after Enter")
	}
	if sel.active {
		t.Error("expected inactive after confirm")
	}

	selected := sel.SelectedItems()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected items, got %d", len(selected))
	}
}

func TestSelectorModel_SingleSelect(t *testing.T) {
	items := []SelectorItem{
		{ID: "1", Name: "v1.0", Selected: true},
		{ID: "2", Name: "v2.0", Selected: false},
	}

	sel := NewSelectorModel("Milestone", items, false)

	// Move to second item
	sel, _ = sel.Update(tea.KeyPressMsg{Code: 'j'})

	// Space in single-select mode deselects all, selects current
	sel, _ = sel.Update(tea.KeyPressMsg{Code: ' '})
	if sel.items[0].Selected {
		t.Error("expected first item to be deselected in single-select")
	}
	if !sel.items[1].Selected {
		t.Error("expected second item to be selected in single-select")
	}
}

func TestSelectorModel_Cancel(t *testing.T) {
	items := []SelectorItem{{ID: "1", Name: "test"}}
	sel := NewSelectorModel("Test", items, false)

	sel, _ = sel.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if !sel.cancelled {
		t.Error("expected cancelled after Esc")
	}
	if sel.active {
		t.Error("expected inactive after cancel")
	}
}

func TestSelectorModel_CursorBounds(t *testing.T) {
	items := []SelectorItem{
		{ID: "1", Name: "a"},
		{ID: "2", Name: "b"},
	}
	sel := NewSelectorModel("Test", items, true)

	// Try to go above 0
	sel, _ = sel.Update(tea.KeyPressMsg{Code: 'k'})
	if sel.cursor != 0 {
		t.Errorf("cursor = %d, should not go below 0", sel.cursor)
	}

	// Go to end
	sel, _ = sel.Update(tea.KeyPressMsg{Code: 'j'})
	sel, _ = sel.Update(tea.KeyPressMsg{Code: 'j'})
	if sel.cursor != 1 {
		t.Errorf("cursor = %d, should not exceed %d", sel.cursor, len(items)-1)
	}
}

func TestSelectorModel_View(t *testing.T) {
	items := []SelectorItem{
		{ID: "1", Name: "bug", Selected: true},
		{ID: "2", Name: "feature", Selected: false},
	}
	sel := NewSelectorModel("Labels", items, true)
	view := sel.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}
