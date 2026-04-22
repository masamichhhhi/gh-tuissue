package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	"github.com/masamichhhhi/gh-tuissue/internal/service"
)

type editType int

const (
	editNone editType = iota
	editTitle
	editBody
	editComment
	editStatus
	editLabels
	editAssignees
	editMilestone
)

// EditingField represents which metadata field is currently being edited inline.
type EditingField int

const (
	EditingNone EditingField = iota
	EditingLabels
	EditingAssignees
	EditingMilestone
)

type DetailModel struct {
	issue    *domain.Issue
	comments []domain.Comment
	content  string
	scroll   int
	width    int
	height   int
	loading  bool
	errorMsg string
	wantEdit editType
	dirty    bool

	// Inline editing
	editingField EditingField
	selector     SelectorModel
	repoSvc      *service.RepoService
	issueSvc     *service.IssueService

	// Metadata cache
	cachedLabels     []domain.Label
	cachedUsers      []domain.User
	cachedMilestones []domain.Milestone
	metadataLoaded   bool
}

func NewDetailModel(repoSvc *service.RepoService, issueSvc *service.IssueService) DetailModel {
	return DetailModel{
		repoSvc:  repoSvc,
		issueSvc: issueSvc,
	}
}

func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *DetailModel) SetIssue(issue domain.Issue) {
	m.issue = &issue
	m.comments = nil
	m.scroll = 0
	m.content = ""
	m.errorMsg = ""
	m.editingField = EditingNone
	m.dirty = false
	m.renderContent()
}

func (m *DetailModel) renderContent() {
	if m.issue == nil {
		return
	}

	var sb strings.Builder
	issue := m.issue

	// Title and status
	sb.WriteString(titleStyle.Render(fmt.Sprintf("#%d %s", issue.Number, issue.Title)))
	sb.WriteString("\n")
	stateColor := "green"
	if issue.State == domain.IssueClosed {
		stateColor = "red"
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(stateColor)).Render(string(issue.State)))
	sb.WriteString("\n\n")

	// Metadata
	if len(issue.Labels) > 0 {
		var labels []string
		for _, l := range issue.Labels {
			color := lipgloss.Color("#" + l.Color)
			labels = append(labels, labelStyle.Background(color).Foreground(lipgloss.Color("0")).Render(l.Name))
		}
		sb.WriteString("Labels: " + strings.Join(labels, " ") + "\n")
	}
	if len(issue.Assignees) > 0 {
		var names []string
		for _, a := range issue.Assignees {
			names = append(names, "@"+a.Login)
		}
		sb.WriteString("Assignees: " + strings.Join(names, ", ") + "\n")
	}
	if issue.Milestone != nil {
		sb.WriteString("Milestone: " + issue.Milestone.Title + "\n")
	}
	sb.WriteString(dimStyle.Render(fmt.Sprintf("Created: %s  Updated: %s",
		issue.CreatedAt.Format(time.DateOnly),
		issue.UpdatedAt.Format(time.DateOnly))) + "\n")
	sb.WriteString("\n")

	// Body (markdown rendered)
	if issue.Body != "" {
		rendered, err := glamour.Render(issue.Body, "dark")
		if err != nil {
			sb.WriteString(issue.Body)
		} else {
			sb.WriteString(rendered)
		}
	} else {
		sb.WriteString(dimStyle.Render("No description provided."))
	}
	sb.WriteString("\n")

	// Comments
	if len(m.comments) > 0 {
		sb.WriteString(titleStyle.Render(fmt.Sprintf("Comments (%d)", len(m.comments))))
		sb.WriteString("\n\n")
		for _, c := range m.comments {
			header := fmt.Sprintf("@%s on %s", c.Author.Login, c.CreatedAt.Format(time.DateOnly))
			sb.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
			sb.WriteString("\n")
			rendered, err := glamour.Render(c.Body, "dark")
			if err != nil {
				sb.WriteString(c.Body)
			} else {
				sb.WriteString(rendered)
			}
			sb.WriteString("\n")
		}
	}

	m.content = sb.String()
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	// When selector is active, route key events to it
	if m.editingField != EditingNone {
		switch msg.(type) {
		case tea.KeyPressMsg:
			m.selector, _ = m.selector.Update(msg)

			if m.selector.confirmed {
				return m.handleSelectorConfirm()
			}
			if m.selector.cancelled {
				m.editingField = EditingNone
				return m, nil
			}
			return m, nil
		case inlineEditMetadataLoadedMsg:
			// Allow metadata to be processed below
		default:
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'j', tea.KeyDown:
			m.scroll++
		case 'k', tea.KeyUp:
			if m.scroll > 0 {
				m.scroll--
			}
		case 's':
			m.wantEdit = editStatus
		case 'e':
			m.wantEdit = editBody
		case 'c':
			m.wantEdit = editComment
		case 't':
			m.wantEdit = editTitle
		case 'l':
			m.wantEdit = editLabels
		case 'a':
			m.wantEdit = editAssignees
		case 'm':
			m.wantEdit = editMilestone
		}
	case commentsLoadedMsg:
		m.comments = msg.comments
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		}
		m.renderContent()
	case inlineEditMetadataLoadedMsg:
		m.cachedLabels = msg.labels
		m.cachedUsers = msg.users
		m.cachedMilestones = msg.milestones
		m.metadataLoaded = true
		// Now populate the selector for the pending edit
		m.populateSelector()
	case inlineEditCompletedMsg:
		m.editingField = EditingNone
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			// Update the issue in place
			m.issue = &msg.issue
			m.dirty = true
			m.renderContent()
		}
	}
	return m, nil
}

func (m *DetailModel) handleSelectorConfirm() (DetailModel, tea.Cmd) {
	if m.issue == nil || m.issueSvc == nil {
		m.editingField = EditingNone
		return *m, nil
	}

	selected := m.selector.SelectedItems()
	number := m.issue.Number
	svc := m.issueSvc
	field := m.editingField
	m.editingField = EditingNone

	switch field {
	case EditingLabels:
		labels := make([]string, len(selected))
		for i, s := range selected {
			labels[i] = s.Name
		}
		return *m, func() tea.Msg {
			issue, err := svc.UpdateIssue(context.Background(), number, service.UpdateIssueInput{
				Labels: &labels,
			})
			return inlineEditCompletedMsg{issue: issue, err: err}
		}
	case EditingAssignees:
		assignees := make([]string, len(selected))
		for i, s := range selected {
			assignees[i] = s.ID
		}
		return *m, func() tea.Msg {
			issue, err := svc.UpdateIssue(context.Background(), number, service.UpdateIssueInput{
				Assignees: &assignees,
			})
			return inlineEditCompletedMsg{issue: issue, err: err}
		}
	case EditingMilestone:
		var milestone string
		if len(selected) > 0 {
			milestone = selected[0].ID
		}
		return *m, func() tea.Msg {
			issue, err := svc.UpdateIssue(context.Background(), number, service.UpdateIssueInput{
				Milestone: &milestone,
			})
			return inlineEditCompletedMsg{issue: issue, err: err}
		}
	}

	return *m, nil
}

func (m *DetailModel) populateSelector() {
	if m.issue == nil {
		return
	}

	switch m.editingField {
	case EditingLabels:
		currentLabels := make(map[string]bool)
		for _, l := range m.issue.Labels {
			currentLabels[l.Name] = true
		}
		items := make([]SelectorItem, len(m.cachedLabels))
		for i, l := range m.cachedLabels {
			items[i] = SelectorItem{
				ID:       l.Name,
				Name:     l.Name,
				Selected: currentLabels[l.Name],
			}
		}
		m.selector = NewSelectorModel("Select Labels", items, true)

	case EditingAssignees:
		currentAssignees := make(map[string]bool)
		for _, a := range m.issue.Assignees {
			currentAssignees[a.Login] = true
		}
		items := make([]SelectorItem, len(m.cachedUsers))
		for i, u := range m.cachedUsers {
			items[i] = SelectorItem{
				ID:       u.Login,
				Name:     "@" + u.Login,
				Selected: currentAssignees[u.Login],
			}
		}
		m.selector = NewSelectorModel("Select Assignees", items, true)

	case EditingMilestone:
		// Add a "(none)" option
		items := make([]SelectorItem, 0, len(m.cachedMilestones)+1)
		currentMilestone := ""
		if m.issue.Milestone != nil {
			currentMilestone = m.issue.Milestone.Title
		}
		items = append(items, SelectorItem{
			ID:       "",
			Name:     "(none)",
			Selected: currentMilestone == "",
		})
		for _, ms := range m.cachedMilestones {
			items = append(items, SelectorItem{
				ID:       fmt.Sprintf("%d", ms.Number),
				Name:     ms.Title,
				Selected: ms.Title == currentMilestone,
			})
		}
		m.selector = NewSelectorModel("Select Milestone", items, false)
	}
}

// StartInlineEdit begins an inline edit for the given field.
// If metadata is not loaded, it returns a command to load it asynchronously.
func (m *DetailModel) StartInlineEdit(field EditingField) tea.Cmd {
	m.editingField = field

	if !m.metadataLoaded {
		if m.repoSvc == nil {
			m.editingField = EditingNone
			return nil
		}
		svc := m.repoSvc
		return func() tea.Msg {
			labels, _ := svc.ListLabels(context.Background())
			users, _ := svc.ListCollaborators(context.Background())
			milestones, _ := svc.ListMilestones(context.Background())
			return inlineEditMetadataLoadedMsg{
				labels:     labels,
				users:      users,
				milestones: milestones,
			}
		}
	}

	// Metadata already loaded, populate immediately
	m.populateSelector()
	return nil
}

func (m DetailModel) View() string {
	if m.issue == nil {
		return "No issue selected"
	}

	// If selector is active, show it as overlay
	if m.editingField != EditingNone && m.selector.active {
		selectorView := m.selector.View()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, selectorView)
	}

	lines := strings.Split(m.content, "\n")
	start := m.scroll
	if start >= len(lines) {
		start = len(lines) - 1
	}
	if start < 0 {
		start = 0
	}
	end := start + m.height
	if end > len(lines) {
		end = len(lines)
	}

	visible := strings.Join(lines[start:end], "\n")

	if m.errorMsg != "" {
		visible = errorStyle.Render(m.errorMsg) + "\n" + visible
	}

	return visible
}

func (m DetailModel) loadComments(svc *service.IssueService) tea.Cmd {
	if svc == nil || m.issue == nil {
		return nil
	}
	number := m.issue.Number
	return func() tea.Msg {
		comments, err := svc.ListComments(context.Background(), number)
		return commentsLoadedMsg{comments: comments, err: err}
	}
}

type commentsLoadedMsg struct {
	comments []domain.Comment
	err      error
}

type inlineEditMetadataLoadedMsg struct {
	labels     []domain.Label
	users      []domain.User
	milestones []domain.Milestone
}

type inlineEditCompletedMsg struct {
	issue domain.Issue
	err   error
}
