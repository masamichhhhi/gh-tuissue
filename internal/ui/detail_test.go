package ui

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

// newTestDetailModel creates a DetailModel with a dummy issue for testing.
func newTestDetailModel() DetailModel {
	m := DetailModel{}
	m.issue = &domain.Issue{
		Number: 1,
		Title:  "Test Issue",
		Labels: []domain.Label{{Name: "bug"}},
	}
	m.width = 80
	m.height = 40
	return m
}

func TestDetailUpdate_MetadataLoadedWhileEditing(t *testing.T) {
	m := newTestDetailModel()

	// Simulate: editingField is already set (user pressed 'l'), waiting for metadata
	m.editingField = EditingLabels

	// Send inlineEditMetadataLoadedMsg while editingField is active
	msg := inlineEditMetadataLoadedMsg{
		labels: []domain.Label{
			{Name: "bug"},
			{Name: "enhancement"},
			{Name: "documentation"},
		},
		users:      []domain.User{{Login: "alice"}},
		milestones: []domain.Milestone{{Number: 1, Title: "v1.0"}},
	}

	m, _ = m.Update(msg)

	// Metadata should be cached
	if !m.metadataLoaded {
		t.Fatal("expected metadataLoaded to be true")
	}
	if len(m.cachedLabels) != 3 {
		t.Errorf("cachedLabels = %d, want 3", len(m.cachedLabels))
	}

	// Selector should be populated and active
	if !m.selector.active {
		t.Fatal("expected selector to be active after metadata loaded")
	}
	if len(m.selector.items) != 3 {
		t.Errorf("selector items = %d, want 3", len(m.selector.items))
	}

	// The current issue label "bug" should be pre-selected
	if !m.selector.items[0].Selected {
		t.Error("expected 'bug' to be pre-selected")
	}
	if m.selector.items[1].Selected {
		t.Error("expected 'enhancement' to not be selected")
	}
}

func TestDetailUpdate_KeyPressIgnoredDuringEditing(t *testing.T) {
	m := newTestDetailModel()
	m.editingField = EditingLabels

	// Non-selector keys should not leak through to detail navigation
	oldScroll := m.scroll
	m, _ = m.Update(tea.KeyPressMsg{Code: 'j'})

	if m.scroll != oldScroll {
		t.Error("expected scroll to not change while editing")
	}
	if m.wantEdit != editNone {
		t.Error("expected wantEdit to remain editNone while editing")
	}
}

func TestDetailUpdate_SelectorCancelRestoresState(t *testing.T) {
	m := newTestDetailModel()
	m.editingField = EditingLabels
	m.metadataLoaded = true
	m.cachedLabels = []domain.Label{{Name: "bug"}, {Name: "fix"}}

	// Populate selector
	m.populateSelector()

	if !m.selector.active {
		t.Fatal("expected selector to be active")
	}

	// Press Escape to cancel
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if m.editingField != EditingNone {
		t.Error("expected editingField to be EditingNone after cancel")
	}
}

func TestDetailUpdate_DirtySetOnInlineEditCompleted(t *testing.T) {
	m := newTestDetailModel()

	if m.dirty {
		t.Fatal("expected dirty to be false initially")
	}

	// Successful inline edit should set dirty
	updatedIssue := *m.issue
	updatedIssue.Labels = []domain.Label{{Name: "bug"}, {Name: "enhancement"}}
	m, _ = m.Update(inlineEditCompletedMsg{issue: updatedIssue, err: nil})

	if !m.dirty {
		t.Error("expected dirty to be true after successful inline edit")
	}
}

func TestDetailUpdate_DirtyNotSetOnInlineEditError(t *testing.T) {
	m := newTestDetailModel()

	m, _ = m.Update(inlineEditCompletedMsg{
		issue: domain.Issue{},
		err:   fmt.Errorf("API error"),
	})

	if m.dirty {
		t.Error("expected dirty to remain false after failed inline edit")
	}
}

func TestDetailUpdate_SetIssueClearsDirty(t *testing.T) {
	m := newTestDetailModel()
	m.dirty = true

	newIssue := domain.Issue{Number: 2, Title: "New Issue"}
	m.SetIssue(newIssue)

	if m.dirty {
		t.Error("expected dirty to be reset after SetIssue")
	}
}

func TestDetailUpdate_LKeySetsWantEdit(t *testing.T) {
	m := newTestDetailModel()

	m, _ = m.Update(tea.KeyPressMsg{Code: 'l'})

	if m.wantEdit != editLabels {
		t.Errorf("wantEdit = %d, want editLabels (%d)", m.wantEdit, editLabels)
	}
}
