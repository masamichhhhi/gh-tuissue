package ui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/masamichhhhi/gh-tuissue/internal/config"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	"github.com/masamichhhhi/gh-tuissue/internal/service"
)

type ViewState int

const (
	ViewBoard ViewState = iota
	ViewDetail
	ViewEditor
	ViewFilter
	ViewHelp
)

type AppModel struct {
	currentView   ViewState
	prevView      ViewState
	board         BoardModel
	detail        DetailModel
	filter        FilterModel
	help          HelpModel
	statusMsg     string
	width         int
	height        int
	issueSvc      *service.IssueService
	repoSvc       *service.RepoService
	projectSvc    *service.ProjectService
	projectNumber int
	projectError  string
	lastEditType  editType
	repoRoot      string
	cfg           *config.Config
}

func NewAppModel(issueSvc *service.IssueService, repoSvc *service.RepoService, projectSvc *service.ProjectService, projectNumber int, repoRoot string, cfg *config.Config) AppModel {
	return AppModel{
		currentView:   ViewBoard,
		board:         NewBoardModel(),
		detail:        NewDetailModel(repoSvc, issueSvc),
		filter:        NewFilterModel(),
		help:          NewHelpModel(),
		issueSvc:      issueSvc,
		repoSvc:       repoSvc,
		projectSvc:    projectSvc,
		projectNumber: projectNumber,
		repoRoot:      repoRoot,
		cfg:           cfg,
	}
}

func (m AppModel) Init() tea.Cmd {
	if m.projectSvc != nil && m.projectNumber > 0 {
		return m.board.loadProjectData(m.projectSvc, m.projectNumber)
	}
	return m.board.loadIssues(m.issueSvc)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.board.SetSize(msg.Width, msg.Height-2)
		m.detail.SetSize(msg.Width, msg.Height-2)
		return m, nil

	case tea.KeyPressMsg:
		// When detail selector is active, delegate all keys to detail
		if m.currentView == ViewDetail && m.detail.editingField != EditingNone {
			var cmd tea.Cmd
			m.detail, cmd = m.detail.Update(msg)
			if m.detail.wantEdit != editNone {
				return m.handleDetailEdit()
			}
			return m, cmd
		}
		// Global keys
		switch {
		case msg.Code == 'c' && msg.Mod.Contains(tea.ModCtrl):
			return m, tea.Quit
		case msg.Code == '?' && m.currentView != ViewEditor:
			m.prevView = m.currentView
			m.currentView = ViewHelp
			return m, nil
		case msg.Code == tea.KeyEscape:
			return m.handleEscape()
		case msg.Code == 'r' && m.currentView == ViewBoard:
			m.board.loading = true
			m.statusMsg = "Refreshing..."
			return m, m.reloadBoardData()
		case msg.Code == 'f' && m.currentView == ViewBoard:
			m.prevView = ViewBoard
			m.currentView = ViewFilter
			return m, m.loadFilterMetadata()
		case msg.Code == 'n' && m.currentView == ViewBoard:
			m.lastEditType = editNone
			return m, launchEditor("# Title\n\nDescription here")
		}

	case projectDataMsg:
		m.board.loading = false
		if msg.err != nil {
			m.projectError = fmt.Sprintf("Project error: %v — falling back to Open/Closed", msg.err)
			m.statusMsg = m.projectError
			return m, m.board.loadIssues(m.issueSvc)
		}
		m.projectError = ""
		m.board.SetProjectData(msg.info, msg.items)
		if m.cfg != nil && len(m.cfg.HiddenColumns) > 0 {
			m.board.ApplyHiddenColumns(m.cfg.HiddenColumns)
		}
		total := 0
		for _, col := range m.board.columns {
			total += len(col.Items)
		}
		m.statusMsg = fmt.Sprintf("%d items loaded from project %q", total, msg.info.Title)
		return m, nil

	case issuesLoadedMsg:
		m.board.loading = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.board.SetFallbackIssues(msg.issues)
			if m.cfg != nil && len(m.cfg.HiddenColumns) > 0 {
				m.board.ApplyHiddenColumns(m.cfg.HiddenColumns)
			}
			if m.projectError != "" {
				m.statusMsg = m.projectError
			} else {
				m.statusMsg = fmt.Sprintf("%d issues loaded", len(msg.issues))
			}
		}
		return m, nil

	case issueUpdatedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.statusMsg = "Issue updated"
			m.currentView = ViewBoard
			return m, m.reloadBoardData()
		}
		return m, nil

	case statusMoveMsg:
		if msg.err != nil {
			m.board.RollbackItemMove(msg.itemID, msg.originalStatus, msg.originalColIdx)
			m.statusMsg = fmt.Sprintf("Status move failed: %v", msg.err)
		} else {
			m.statusMsg = "Status updated"
		}
		return m, nil

	case editorResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Editor error: %v", msg.err)
			m.currentView = ViewBoard
			return m, nil
		}
		if m.lastEditType == editNone {
			return m.handleNewIssue(msg.content)
		}
		return m.handleEditorResult(msg.content)

	case filterMetadataMsg:
		m.filter.SetMetadata(msg.labels, msg.users, msg.milestones)
		return m, nil

	case inlineEditCompletedMsg:
		m.detail, _ = m.detail.Update(msg)
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.statusMsg = "Issue updated"
		}
		return m, nil

	case inlineEditMetadataLoadedMsg:
		m.detail, _ = m.detail.Update(msg)
		return m, nil

	case statusMsg:
		m.statusMsg = string(msg)
		return m, nil
	}

	// Delegate to current view
	var cmd tea.Cmd
	switch m.currentView {
	case ViewBoard:
		var boardMsg tea.Msg = msg
		m.board, cmd = m.board.Update(boardMsg)
		if m.board.wantStatusMsg != "" {
			m.statusMsg = m.board.wantStatusMsg
			m.board.wantStatusMsg = ""
		}
		if m.board.wantConfigUpdate {
			m.board.wantConfigUpdate = false
			m.saveHiddenColumns()
		}
		if m.board.selectedIssue != nil {
			m.detail.SetIssue(*m.board.selectedIssue)
			m.board.selectedIssue = nil
			m.prevView = ViewBoard
			m.currentView = ViewDetail
			return m, m.detail.loadComments(m.issueSvc)
		}
		if m.board.wantStatusMove != 0 {
			return m.handleStatusMove()
		}
	case ViewDetail:
		m.detail, cmd = m.detail.Update(msg)
		if m.detail.wantEdit != editNone {
			return m.handleDetailEdit()
		}
	case ViewFilter:
		m.filter, cmd = m.filter.Update(msg)
		if m.filter.applied {
			m.filter.applied = false
			m.board.SetFilter(m.filter.State())
			m.currentView = ViewBoard
			return m, nil
		}
	case ViewHelp:
		m.help, cmd = m.help.Update(msg)
	}

	return m, cmd
}

func (m AppModel) reloadBoardData() tea.Cmd {
	if m.projectSvc != nil && m.projectNumber > 0 {
		return m.board.loadProjectData(m.projectSvc, m.projectNumber)
	}
	return m.board.loadIssues(m.issueSvc)
}

func (m AppModel) handleStatusMove() (tea.Model, tea.Cmd) {
	direction := m.board.wantStatusMove
	m.board.wantStatusMove = 0

	if m.projectSvc == nil || m.board.projectInfo == nil {
		m.statusMsg = "Status move not available (no project)"
		return m, nil
	}

	item := m.board.SelectedItem()
	if item == nil {
		return m, nil
	}

	options := m.board.projectInfo.StatusField.Options
	currentIdx := -1
	for i, opt := range options {
		if opt.ID == item.StatusID {
			currentIdx = i
			break
		}
	}

	targetIdx := currentIdx + direction
	if targetIdx < 0 || targetIdx >= len(options) {
		return m, nil
	}

	// Save rollback info before optimistic update
	itemID := item.ItemID
	originalStatus := item.StatusID
	originalColIdx := currentIdx

	// Optimistic UI update: move item locally before API call
	m.board.MoveItemToColumn(itemID, currentIdx, targetIdx)

	targetOption := options[targetIdx]
	projectID := m.board.projectInfo.ID
	fieldID := m.board.projectInfo.StatusField.ID
	optionID := targetOption.ID
	svc := m.projectSvc

	return m, func() tea.Msg {
		err := svc.MoveItemStatus(context.Background(), projectID, itemID, fieldID, optionID)
		return statusMoveMsg{
			err:            err,
			itemID:         itemID,
			originalStatus: originalStatus,
			originalColIdx: originalColIdx,
		}
	}
}

func (m *AppModel) saveHiddenColumns() {
	if m.repoRoot == "" {
		return
	}
	if m.cfg == nil {
		m.cfg = &config.Config{}
	}
	m.cfg.HiddenColumns = m.board.HiddenColumnNames()
	_ = config.Save(m.repoRoot, *m.cfg)
}

func (m AppModel) handleEscape() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ViewHelp, ViewFilter:
		m.currentView = m.prevView
	case ViewDetail:
		m.currentView = ViewBoard
	case ViewBoard:
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) handleDetailEdit() (tea.Model, tea.Cmd) {
	edit := m.detail.wantEdit
	m.detail.wantEdit = editNone

	if m.detail.issue == nil || m.issueSvc == nil {
		return m, nil
	}
	issue := m.detail.issue

	switch edit {
	case editStatus:
		return m, m.toggleIssueStatus(issue)
	case editBody:
		m.lastEditType = editBody
		return m, launchEditor(issue.Body)
	case editComment:
		m.lastEditType = editComment
		return m, launchEditor("")
	case editTitle:
		m.lastEditType = editTitle
		return m, launchEditor(issue.Title)
	case editLabels:
		cmd := m.detail.StartInlineEdit(EditingLabels)
		return m, cmd
	case editAssignees:
		cmd := m.detail.StartInlineEdit(EditingAssignees)
		return m, cmd
	case editMilestone:
		cmd := m.detail.StartInlineEdit(EditingMilestone)
		return m, cmd
	}
	return m, nil
}

func (m AppModel) toggleIssueStatus(issue *domain.Issue) tea.Cmd {
	svc := m.issueSvc
	number := issue.Number
	isOpen := issue.State == domain.IssueOpen
	return func() tea.Msg {
		var err error
		if isOpen {
			_, err = svc.CloseIssue(context.Background(), number)
		} else {
			_, err = svc.ReopenIssue(context.Background(), number)
		}
		return issueUpdatedMsg{err: err}
	}
}

func (m AppModel) loadFilterMetadata() tea.Cmd {
	if m.repoSvc == nil {
		return nil
	}
	svc := m.repoSvc
	return func() tea.Msg {
		labels, _ := svc.ListLabels(context.Background())
		users, _ := svc.ListCollaborators(context.Background())
		milestones, _ := svc.ListMilestones(context.Background())
		return filterMetadataMsg{labels: labels, users: users, milestones: milestones}
	}
}

type filterMetadataMsg struct {
	labels     []domain.Label
	users      []domain.User
	milestones []domain.Milestone
}

func (m AppModel) handleEditorResult(content string) (tea.Model, tea.Cmd) {
	if m.detail.issue == nil || m.issueSvc == nil {
		m.currentView = ViewBoard
		return m, nil
	}

	svc := m.issueSvc
	number := m.detail.issue.Number
	lastEdit := m.lastEditType
	// Return to board after edit completes (5.8)
	m.currentView = ViewBoard

	switch lastEdit {
	case editComment:
		return m, func() tea.Msg {
			_, err := svc.AddComment(context.Background(), number, content)
			return issueUpdatedMsg{err: err}
		}
	case editTitle:
		return m, func() tea.Msg {
			title := strings.TrimSpace(content)
			_, err := svc.UpdateIssue(context.Background(), number, service.UpdateIssueInput{
				Title: &title,
			})
			return issueUpdatedMsg{err: err}
		}
	default:
		return m, func() tea.Msg {
			body := content
			_, err := svc.UpdateIssue(context.Background(), number, service.UpdateIssueInput{
				Body: &body,
			})
			return issueUpdatedMsg{err: err}
		}
	}
}

func (m AppModel) handleNewIssue(content string) (tea.Model, tea.Cmd) {
	if m.issueSvc == nil {
		return m, nil
	}

	lines := strings.SplitN(content, "\n", 2)
	title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "#"))
	if title == "" || title == "Title" {
		m.statusMsg = "Issue creation cancelled (no title)"
		return m, nil
	}
	var body string
	if len(lines) > 1 {
		body = strings.TrimSpace(lines[1])
		if body == "Description here" {
			body = ""
		}
	}

	svc := m.issueSvc
	return m, func() tea.Msg {
		_, err := svc.CreateIssue(context.Background(), service.CreateIssueInput{
			Title: title,
			Body:  body,
		})
		if err != nil {
			return issueUpdatedMsg{err: err}
		}
		return issueUpdatedMsg{err: nil}
	}
}

func (m AppModel) View() tea.View {
	var content string

	switch m.currentView {
	case ViewBoard:
		content = m.board.View()
	case ViewDetail:
		content = m.detail.View()
	case ViewFilter:
		content = m.filter.View()
	case ViewHelp:
		content = m.help.View()
	default:
		content = m.board.View()
	}

	// Status bar
	hints := m.keyHints()
	status := m.statusMsg
	if status == "" {
		status = " "
	}
	statusBar := statusBarStyle.Width(m.width).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			status,
			lipgloss.NewStyle().Width(m.width-lipgloss.Width(status)-lipgloss.Width(hints)).Render(""),
			hints,
		),
	)

	view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, content, statusBar))
	view.AltScreen = true
	return view
}

func (m AppModel) keyHints() string {
	switch m.currentView {
	case ViewBoard:
		return dimStyle.Render("h/l/←/→:column j/k/↑/↓:move H/L:status d:hide D:show enter:open f:filter r:refresh ?:help")
	case ViewDetail:
		return dimStyle.Render("j/k/↑/↓:scroll s:status l:labels a:assign m:milestone e:edit c:comment esc:back")
	case ViewFilter:
		return dimStyle.Render("j/k:move space:toggle enter:apply esc:cancel")
	case ViewHelp:
		return dimStyle.Render("esc:close")
	default:
		return ""
	}
}

// Messages
type issuesLoadedMsg struct {
	issues []issueWithState
	err    error
}

type issueUpdatedMsg struct {
	err error
}

type statusMsg string
