package domain

import (
	"errors"
	"testing"
	"time"
)

func TestIssueState(t *testing.T) {
	t.Run("valid states", func(t *testing.T) {
		if IssueOpen != "OPEN" {
			t.Errorf("IssueOpen = %q, want %q", IssueOpen, "OPEN")
		}
		if IssueClosed != "CLOSED" {
			t.Errorf("IssueClosed = %q, want %q", IssueClosed, "CLOSED")
		}
	})

	t.Run("IsOpen", func(t *testing.T) {
		open := Issue{State: IssueOpen}
		if !open.IsOpen() {
			t.Error("Issue with OPEN state should return true for IsOpen()")
		}
		closed := Issue{State: IssueClosed}
		if closed.IsOpen() {
			t.Error("Issue with CLOSED state should return false for IsOpen()")
		}
	})
}

func TestIssueFields(t *testing.T) {
	now := time.Now()
	milestone := Milestone{Number: 1, Title: "v1.0", State: "open", DueOn: &now}
	issue := Issue{
		Number:    42,
		NodeID:    "I_abc123",
		Title:     "Test Issue",
		Body:      "Body content",
		State:     IssueOpen,
		URL:       "https://github.com/owner/repo/issues/42",
		Author:    User{Login: "author", Name: "Author Name"},
		Labels:    []Label{{Name: "bug", Color: "d73a4a", Description: "Something isn't working"}},
		Assignees: []User{{Login: "dev1", Name: "Developer"}},
		Milestone: &milestone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if issue.Number != 42 {
		t.Errorf("Number = %d, want 42", issue.Number)
	}
	if issue.NodeID != "I_abc123" {
		t.Errorf("NodeID = %q, want %q", issue.NodeID, "I_abc123")
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Title = %q, want %q", issue.Title, "Test Issue")
	}
	if len(issue.Labels) != 1 || issue.Labels[0].Name != "bug" {
		t.Errorf("Labels = %+v, want [{Name:bug}]", issue.Labels)
	}
	if issue.Milestone == nil || issue.Milestone.Title != "v1.0" {
		t.Errorf("Milestone = %+v, want v1.0", issue.Milestone)
	}
}

func TestProjectTypes(t *testing.T) {
	t.Run("ProjectSummary", func(t *testing.T) {
		ps := ProjectSummary{ID: "PVT_123", Number: 1, Title: "My Project"}
		if ps.ID != "PVT_123" {
			t.Errorf("ID = %q, want %q", ps.ID, "PVT_123")
		}
		if ps.Number != 1 {
			t.Errorf("Number = %d, want 1", ps.Number)
		}
	})

	t.Run("ProjectInfo with StatusField", func(t *testing.T) {
		pi := ProjectInfo{
			ID:    "PVT_123",
			Title: "My Project",
			StatusField: StatusField{
				ID:   "PVTSSF_456",
				Name: "Status",
				Options: []StatusOption{
					{ID: "opt1", Name: "Todo"},
					{ID: "opt2", Name: "In Progress"},
					{ID: "opt3", Name: "Done"},
				},
			},
		}
		if len(pi.StatusField.Options) != 3 {
			t.Errorf("StatusField.Options length = %d, want 3", len(pi.StatusField.Options))
		}
		if pi.StatusField.Options[0].Name != "Todo" {
			t.Errorf("First option = %q, want %q", pi.StatusField.Options[0].Name, "Todo")
		}
	})

	t.Run("ProjectItem", func(t *testing.T) {
		item := ProjectItem{
			ItemID:   "PVTI_789",
			Issue:    Issue{Number: 42, NodeID: "I_abc123", Title: "Test"},
			StatusID: "opt1",
		}
		if item.ItemID != "PVTI_789" {
			t.Errorf("ItemID = %q, want %q", item.ItemID, "PVTI_789")
		}
		if item.Issue.Number != 42 {
			t.Errorf("Issue.Number = %d, want 42", item.Issue.Number)
		}
	})
}

func TestPageInfo(t *testing.T) {
	p := PageInfo{HasNextPage: true, EndCursor: "abc123"}
	if !p.HasNextPage {
		t.Error("HasNextPage should be true")
	}
	if p.EndCursor != "abc123" {
		t.Errorf("EndCursor = %q, want %q", p.EndCursor, "abc123")
	}
}

func TestAppError(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		appErr := &AppError{
			Code:    ErrAuth,
			Message: "authentication failed",
			Err:     errors.New("token expired"),
		}
		got := appErr.Error()
		if got != "authentication failed: token expired" {
			t.Errorf("Error() = %q, want %q", got, "authentication failed: token expired")
		}
	})

	t.Run("Error without wrapped error", func(t *testing.T) {
		appErr := &AppError{
			Code:    ErrNotFound,
			Message: "issue not found",
		}
		got := appErr.Error()
		if got != "issue not found" {
			t.Errorf("Error() = %q, want %q", got, "issue not found")
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		inner := errors.New("inner error")
		appErr := &AppError{
			Code:    ErrNetwork,
			Message: "network failure",
			Err:     inner,
		}
		if !errors.Is(appErr, inner) {
			t.Error("errors.Is should match wrapped error")
		}
	})

	t.Run("error codes", func(t *testing.T) {
		codes := []ErrorCode{ErrAuth, ErrNetwork, ErrNotFound, ErrPermission, ErrValidation, ErrRateLimit, ErrProjectNotFound, ErrStatusFieldMissing, ErrUnknown}
		for i, code := range codes {
			if int(code) != i {
				t.Errorf("ErrorCode %d has value %d, want %d", i, int(code), i)
			}
		}
	})
}
