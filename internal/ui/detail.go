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
}

func NewDetailModel() DetailModel {
	return DetailModel{}
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
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'j':
			m.scroll++
		case 'k':
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
	}
	return m, nil
}

func (m DetailModel) View() string {
	if m.issue == nil {
		return "No issue selected"
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
