package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// SelectorItem represents a selectable item in the selector.
type SelectorItem struct {
	ID       string
	Name     string
	Selected bool
}

// SelectorModel is a generic select UI component that supports
// both multi-select (labels, assignees) and single-select (milestone).
type SelectorModel struct {
	items       []SelectorItem
	cursor      int
	multiSelect bool
	title       string
	active      bool
	confirmed   bool
	cancelled   bool
}

// NewSelectorModel creates a new SelectorModel.
func NewSelectorModel(title string, items []SelectorItem, multiSelect bool) SelectorModel {
	return SelectorModel{
		items:       items,
		cursor:      0,
		multiSelect: multiSelect,
		title:       title,
		active:      true,
	}
}

// SelectedItems returns items where Selected is true.
func (m SelectorModel) SelectedItems() []SelectorItem {
	var selected []SelectorItem
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// Update handles key events for the selector.
func (m SelectorModel) Update(msg tea.Msg) (SelectorModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'j', tea.KeyDown:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case 'k', tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case ' ':
			if m.cursor < len(m.items) {
				if m.multiSelect {
					m.items[m.cursor].Selected = !m.items[m.cursor].Selected
				} else {
					// Single-select: deselect all, then select current
					for i := range m.items {
						m.items[i].Selected = false
					}
					m.items[m.cursor].Selected = true
				}
			}
		case tea.KeyEnter:
			m.confirmed = true
			m.active = false
		case tea.KeyEscape:
			m.cancelled = true
			m.active = false
		}
	}

	return m, nil
}

// View renders the selector list with cursor indicator and checkmarks.
func (m SelectorModel) View() string {
	var sb strings.Builder

	titleRendered := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Render(m.title)
	sb.WriteString(titleRendered)
	sb.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		check := "[ ] "
		if item.Selected {
			check = "[x] "
		}

		line := fmt.Sprintf("%s%s%s", cursor, check, item.Name)
		if i == m.cursor {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(line)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	if m.multiSelect {
		sb.WriteString(dimStyle.Render("j/k:move space:toggle enter:confirm esc:cancel"))
	} else {
		sb.WriteString(dimStyle.Render("j/k:move space:select enter:confirm esc:cancel"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Padding(1, 2).
		Render(sb.String())
}
