package ui

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

// datedItem returns a Todo item whose issue was created and updated the given
// number of days after a fixed base date.
func datedItem(number, createdDay, updatedDay int) domain.ProjectItem {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return domain.ProjectItem{
		ItemID: fmt.Sprintf("PVTI_%d", number),
		Issue: domain.Issue{
			Number:    number,
			Title:     fmt.Sprintf("Issue %d", number),
			State:     domain.IssueOpen,
			CreatedAt: base.AddDate(0, 0, createdDay),
			UpdatedAt: base.AddDate(0, 0, updatedDay),
		},
		StatusID: "opt_todo",
	}
}

// datedItems returns items loaded in the order #2, #1, #3. #1 was created
// first but updated last; #3 was created last.
func datedItems() []domain.ProjectItem {
	return []domain.ProjectItem{
		datedItem(2, 10, 10),
		datedItem(1, 0, 30),
		datedItem(3, 20, 15),
	}
}

func issueNumbers(items []domain.ProjectItem) []int {
	numbers := make([]int, len(items))
	for i, item := range items {
		numbers[i] = item.Issue.Number
	}
	return numbers
}

func TestParseSortOrder(t *testing.T) {
	for _, s := range []string{"", "created-desc", "created-asc", "updated-desc", "updated-asc"} {
		got, ok := ParseSortOrder(s)
		if !ok || string(got) != s {
			t.Errorf("ParseSortOrder(%q) = %q, %v; want %q, true", s, got, ok, s)
		}
	}
	if got, ok := ParseSortOrder("newest"); ok || got != SortDefault {
		t.Errorf("ParseSortOrder(\"newest\") = %q, %v; want default, false", got, ok)
	}
}

func TestSortItems(t *testing.T) {
	tests := []struct {
		order SortOrder
		want  []int
	}{
		{SortDefault, []int{2, 1, 3}},
		{SortCreatedDesc, []int{3, 2, 1}},
		{SortCreatedAsc, []int{1, 2, 3}},
		{SortUpdatedDesc, []int{1, 3, 2}},
		{SortUpdatedAsc, []int{2, 3, 1}},
	}
	for _, tt := range tests {
		items := datedItems()
		sortItems(items, tt.order)
		if got := issueNumbers(items); !slices.Equal(got, tt.want) {
			t.Errorf("sortItems(%q) = %v, want %v", tt.order, got, tt.want)
		}
	}
}

func TestSortItems_TieBreaksOnNumber(t *testing.T) {
	items := []domain.ProjectItem{datedItem(2, 5, 5), datedItem(1, 5, 5), datedItem(3, 5, 5)}

	sortItems(items, SortCreatedDesc)
	if got := issueNumbers(items); !slices.Equal(got, []int{3, 2, 1}) {
		t.Errorf("desc tie-break = %v, want [3 2 1]", got)
	}
	sortItems(items, SortUpdatedAsc)
	if got := issueNumbers(items); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("asc tie-break = %v, want [1 2 3]", got)
	}
}

func TestSortOrder_Description(t *testing.T) {
	if got := SortCreatedDesc.Description(); got != "created ↓" {
		t.Errorf("Description() = %q, want %q", got, "created ↓")
	}
	if got := SortDefault.Description(); got != "" {
		t.Errorf("default Description() = %q, want empty", got)
	}
}

func TestNewSortPicker_StartsOnCurrentOrder(t *testing.T) {
	picker := newSortPicker(SortUpdatedDesc)
	selected := picker.SelectedItems()
	if len(selected) != 1 || selected[0].ID != string(SortUpdatedDesc) {
		t.Fatalf("selected = %+v, want only %q", selected, SortUpdatedDesc)
	}
	if picker.cursor != 3 {
		t.Errorf("cursor = %d, want 3 (the current order)", picker.cursor)
	}
}
