package ui

import (
	"cmp"
	"slices"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

// SortOrder is the order of cards within every board column. Its value is
// also what is stored in the config file.
type SortOrder string

const (
	// SortDefault keeps the order the items were loaded in.
	SortDefault     SortOrder = ""
	SortCreatedDesc SortOrder = "created-desc"
	SortCreatedAsc  SortOrder = "created-asc"
	SortUpdatedDesc SortOrder = "updated-desc"
	SortUpdatedAsc  SortOrder = "updated-asc"
)

// sortOptions lists the orders offered by the sort picker, in display order.
var sortOptions = []struct {
	order SortOrder
	label string
}{
	{SortDefault, "Default order"},
	{SortCreatedDesc, "Created: newest first"},
	{SortCreatedAsc, "Created: oldest first"},
	{SortUpdatedDesc, "Updated: newest first"},
	{SortUpdatedAsc, "Updated: oldest first"},
}

// ParseSortOrder validates a sort order read from the config file.
func ParseSortOrder(s string) (SortOrder, bool) {
	for _, opt := range sortOptions {
		if string(opt.order) == s {
			return opt.order, true
		}
	}
	return SortDefault, false
}

// Description renders the order for the board header, e.g. "created ↓".
func (o SortOrder) Description() string {
	switch o {
	case SortCreatedDesc:
		return "created ↓"
	case SortCreatedAsc:
		return "created ↑"
	case SortUpdatedDesc:
		return "updated ↓"
	case SortUpdatedAsc:
		return "updated ↑"
	default:
		return ""
	}
}

// sortItems sorts items in place. SortDefault leaves them untouched; ties are
// broken by issue number in the same direction.
func sortItems(items []domain.ProjectItem, o SortOrder) {
	if o == SortDefault {
		return
	}
	byUpdated := o == SortUpdatedDesc || o == SortUpdatedAsc
	desc := o == SortCreatedDesc || o == SortUpdatedDesc
	slices.SortStableFunc(items, func(a, b domain.ProjectItem) int {
		c := a.Issue.CreatedAt.Compare(b.Issue.CreatedAt)
		if byUpdated {
			c = a.Issue.UpdatedAt.Compare(b.Issue.UpdatedAt)
		}
		if c == 0 {
			c = cmp.Compare(a.Issue.Number, b.Issue.Number)
		}
		if desc {
			return -c
		}
		return c
	})
}

// newSortPicker builds the sort picker with the current order pre-selected.
func newSortPicker(current SortOrder) SelectorModel {
	items := make([]SelectorItem, len(sortOptions))
	for i, opt := range sortOptions {
		items[i] = SelectorItem{ID: string(opt.order), Name: opt.label, Selected: opt.order == current}
	}
	return NewSelectorModel("Sort issues", items, false).WithEnterSelects()
}
