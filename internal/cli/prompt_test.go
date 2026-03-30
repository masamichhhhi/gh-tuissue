package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/masamichhhhi/gh-tuissue/internal/config"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

type mockProjectLister struct {
	projects []domain.ProjectSummary
	err      error
}

func (m *mockProjectLister) ListProjects(_ context.Context) ([]domain.ProjectSummary, error) {
	return m.projects, m.err
}

func TestPromptProjectSelection_UserSelectsNo(t *testing.T) {
	lister := &mockProjectLister{}
	confirm := func(title string) (bool, error) { return false, nil }
	sel := func(title string, options []selectOption) (int, error) {
		t.Fatal("select should not be called when user says No")
		return 0, nil
	}

	result, err := promptProjectSelectionWith(lister, "", confirm, sel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestPromptProjectSelection_UserSelectsYesAndPicksProject(t *testing.T) {
	lister := &mockProjectLister{
		projects: []domain.ProjectSummary{
			{ID: "1", Number: 1, Title: "Alpha"},
			{ID: "2", Number: 2, Title: "Beta"},
		},
	}
	confirm := func(title string) (bool, error) { return true, nil }
	sel := func(title string, options []selectOption) (int, error) {
		if len(options) != 2 {
			t.Fatalf("expected 2 options, got %d", len(options))
		}
		return options[1].Value, nil // pick Beta (#2)
	}

	tmpDir := t.TempDir()
	result, err := promptProjectSelectionWith(lister, tmpDir, confirm, sel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 2 {
		t.Errorf("expected 2, got %d", result)
	}

	// Verify config was saved
	cfg, err := config.Load(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg == nil || cfg.ProjectNumber != 2 {
		t.Errorf("config not saved correctly: %v", cfg)
	}
}

func TestPromptProjectSelection_NoProjectsFound(t *testing.T) {
	lister := &mockProjectLister{projects: nil}
	confirm := func(title string) (bool, error) { return true, nil }
	sel := func(title string, options []selectOption) (int, error) {
		t.Fatal("select should not be called when no projects")
		return 0, nil
	}

	result, err := promptProjectSelectionWith(lister, "", confirm, sel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestPromptProjectSelection_ListProjectsError(t *testing.T) {
	lister := &mockProjectLister{err: fmt.Errorf("API error")}
	confirm := func(title string) (bool, error) { return true, nil }
	sel := func(title string, options []selectOption) (int, error) { return 0, nil }

	_, err := promptProjectSelectionWith(lister, "", confirm, sel)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPromptProjectSelection_ConfirmCancelled(t *testing.T) {
	lister := &mockProjectLister{}
	confirm := func(title string) (bool, error) { return false, fmt.Errorf("cancelled") }
	sel := func(title string, options []selectOption) (int, error) { return 0, nil }

	_, err := promptProjectSelectionWith(lister, "", confirm, sel)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPromptProjectSelection_SelectCancelled(t *testing.T) {
	lister := &mockProjectLister{
		projects: []domain.ProjectSummary{{ID: "1", Number: 1, Title: "Test"}},
	}
	confirm := func(title string) (bool, error) { return true, nil }
	sel := func(title string, options []selectOption) (int, error) {
		return 0, fmt.Errorf("cancelled")
	}

	_, err := promptProjectSelectionWith(lister, "", confirm, sel)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPromptProjectSelection_NoRepoRoot(t *testing.T) {
	lister := &mockProjectLister{
		projects: []domain.ProjectSummary{{ID: "1", Number: 5, Title: "Test"}},
	}
	confirm := func(title string) (bool, error) { return true, nil }
	sel := func(title string, options []selectOption) (int, error) {
		return 5, nil
	}

	// repoRoot is empty, so config should not be saved (no error)
	result, err := promptProjectSelectionWith(lister, "", confirm, sel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}

	// Verify no config file created in current dir
	_, err = os.Stat(filepath.Join(".", config.FileName))
	if err == nil {
		t.Error("config file should not have been created")
	}
}
