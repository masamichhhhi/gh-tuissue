package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

type FilterPane int

const (
	PaneLabels FilterPane = iota
	PaneAssignees
	PaneMilestone
	PaneSort
)

type FilterModel struct {
	state           FilterState
	availLabels     []domain.Label
	availUsers      []domain.User
	availMilestones []domain.Milestone
	activePane      FilterPane
	cursor          int
	applied         bool
	selectedLabels  map[string]bool
	selectedUsers   map[string]bool
}

func NewFilterModel() FilterModel {
	return FilterModel{
		selectedLabels: make(map[string]bool),
		selectedUsers:  make(map[string]bool),
	}
}

func (m *FilterModel) SetMetadata(labels []domain.Label, users []domain.User, milestones []domain.Milestone) {
	m.availLabels = labels
	m.availUsers = users
	m.availMilestones = milestones
}

func (m FilterModel) State() FilterState {
	var labels []string
	for l, selected := range m.selectedLabels {
		if selected {
			labels = append(labels, l)
		}
	}
	var assignees []string
	for a, selected := range m.selectedUsers {
		if selected {
			assignees = append(assignees, a)
		}
	}
	return FilterState{
		Labels:    labels,
		Assignees: assignees,
		Milestone: m.state.Milestone,
	}
}

func (m FilterModel) currentListLen() int {
	switch m.activePane {
	case PaneLabels:
		return len(m.availLabels)
	case PaneAssignees:
		return len(m.availUsers)
	case PaneMilestone:
		return len(m.availMilestones) + 1 // +1 for "none"
	default:
		return 0
	}
}

func (m FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'j', tea.KeyDown:
			max := m.currentListLen()
			if m.cursor < max-1 {
				m.cursor++
			}
		case 'k', tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyTab:
			m.activePane = (m.activePane + 1) % 3
			m.cursor = 0
		case ' ':
			m.toggleSelection()
		case tea.KeyEnter:
			m.applied = true
		}
	}
	return m, nil
}

func (m *FilterModel) toggleSelection() {
	switch m.activePane {
	case PaneLabels:
		if m.cursor < len(m.availLabels) {
			name := m.availLabels[m.cursor].Name
			m.selectedLabels[name] = !m.selectedLabels[name]
		}
	case PaneAssignees:
		if m.cursor < len(m.availUsers) {
			login := m.availUsers[m.cursor].Login
			m.selectedUsers[login] = !m.selectedUsers[login]
		}
	case PaneMilestone:
		if m.cursor == 0 {
			m.state.Milestone = ""
		} else if m.cursor-1 < len(m.availMilestones) {
			m.state.Milestone = m.availMilestones[m.cursor-1].Title
		}
	}
}

func (m FilterModel) View() string {
	var sections []string

	// Labels
	labelsTitle := "Labels"
	if m.activePane == PaneLabels {
		labelsTitle = "▸ " + labelsTitle
	}
	sections = append(sections, titleStyle.Render(labelsTitle))
	for i, l := range m.availLabels {
		check := "  "
		if m.selectedLabels[l.Name] {
			check = "✓ "
		}
		cursor := "  "
		if m.activePane == PaneLabels && m.cursor == i {
			cursor = "> "
		}
		sections = append(sections, fmt.Sprintf("%s%s%s", cursor, check, l.Name))
	}

	sections = append(sections, "")

	// Assignees
	assigneesTitle := "Assignees"
	if m.activePane == PaneAssignees {
		assigneesTitle = "▸ " + assigneesTitle
	}
	sections = append(sections, titleStyle.Render(assigneesTitle))
	for i, u := range m.availUsers {
		check := "  "
		if m.selectedUsers[u.Login] {
			check = "✓ "
		}
		cursor := "  "
		if m.activePane == PaneAssignees && m.cursor == i {
			cursor = "> "
		}
		sections = append(sections, fmt.Sprintf("%s%s@%s", cursor, check, u.Login))
	}

	sections = append(sections, "")

	// Milestones
	msTitle := "Milestone"
	if m.activePane == PaneMilestone {
		msTitle = "▸ " + msTitle
	}
	sections = append(sections, titleStyle.Render(msTitle))
	{
		cursor := "  "
		check := "  "
		if m.state.Milestone == "" {
			check = "✓ "
		}
		if m.activePane == PaneMilestone && m.cursor == 0 {
			cursor = "> "
		}
		sections = append(sections, fmt.Sprintf("%s%s(none)", cursor, check))
	}
	for i, ms := range m.availMilestones {
		cursor := "  "
		check := "  "
		if m.state.Milestone == ms.Title {
			check = "✓ "
		}
		if m.activePane == PaneMilestone && m.cursor == i+1 {
			cursor = "> "
		}
		sections = append(sections, fmt.Sprintf("%s%s%s", cursor, check, ms.Title))
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(sections, "\n"))
}
