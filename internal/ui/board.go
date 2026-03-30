package ui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	"github.com/masamichhhhi/gh-tuissue/internal/service"
)

type issueWithState = domain.Issue

type FilterState struct {
	Labels    []string
	Assignees []string
	Milestone string
	SortField string
	SortDir   string
}

// StatusColumn represents a single column in the kanban board.
type StatusColumn struct {
	OptionID string
	Name     string
	Items    []domain.ProjectItem
}

type BoardModel struct {
	columns       []StatusColumn
	allItems      []domain.ProjectItem
	activeCol     int
	cursorIndex   map[int]int
	scrollOffset  map[int]int
	loading       bool
	filterState   FilterState
	selectedIssue *domain.Issue
	projectInfo   *domain.ProjectInfo
	width         int
	height        int
	// Status move result
	wantStatusMove int // -1=left, 1=right, 0=none
	// Column visibility
	hiddenCols     map[int]bool
	wantStatusMsg  string
}

func NewBoardModel() BoardModel {
	return BoardModel{
		cursorIndex:  make(map[int]int),
		scrollOffset: make(map[int]int),
		hiddenCols:   make(map[int]bool),
		loading:      true,
	}
}

// visibleColumns returns the real indices of columns that are not hidden.
func (m BoardModel) visibleColumns() []int {
	var vis []int
	for i := range m.columns {
		if !m.hiddenCols[i] {
			vis = append(vis, i)
		}
	}
	return vis
}

// activeToReal converts the current activeCol (which is a real index) to a real index.
// This is an identity since activeCol always stores the real index.
func (m BoardModel) activeToReal() int {
	return m.activeCol
}

// realToActive returns the position of realIdx among visible columns (0-based).
func (m BoardModel) realToActive(realIdx int) int {
	pos := 0
	for i := range m.columns {
		if m.hiddenCols[i] {
			continue
		}
		if i == realIdx {
			return pos
		}
		pos++
	}
	return 0
}

// HiddenCount returns the number of hidden columns.
func (m BoardModel) HiddenCount() int {
	return len(m.hiddenCols)
}

func (m *BoardModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// SetProjectData sets the project info and items, building columns from status options.
func (m *BoardModel) SetProjectData(info domain.ProjectInfo, items []domain.ProjectItem) {
	m.projectInfo = &info
	m.allItems = items

	// Build columns from status field options
	m.columns = make([]StatusColumn, len(info.StatusField.Options))
	for i, opt := range info.StatusField.Options {
		m.columns[i] = StatusColumn{
			OptionID: opt.ID,
			Name:     opt.Name,
		}
	}

	m.applyFilter()

	// Initialize cursor/scroll for each column
	for i := range m.columns {
		if _, ok := m.cursorIndex[i]; !ok {
			m.cursorIndex[i] = 0
		}
		if _, ok := m.scrollOffset[i]; !ok {
			m.scrollOffset[i] = 0
		}
	}
}

// SetFallbackIssues sets issues without project context (Open/Closed fallback).
func (m *BoardModel) SetFallbackIssues(issues []domain.Issue) {
	m.projectInfo = nil
	m.columns = []StatusColumn{
		{OptionID: "open", Name: "Open"},
		{OptionID: "closed", Name: "Closed"},
	}

	// Convert issues to ProjectItems
	m.allItems = make([]domain.ProjectItem, len(issues))
	for i, issue := range issues {
		statusID := "open"
		if issue.State == domain.IssueClosed {
			statusID = "closed"
		}
		m.allItems[i] = domain.ProjectItem{
			ItemID:   fmt.Sprintf("fallback-%d", issue.Number),
			Issue:    issue,
			StatusID: statusID,
		}
	}

	m.applyFilter()

	for i := range m.columns {
		if _, ok := m.cursorIndex[i]; !ok {
			m.cursorIndex[i] = 0
		}
		if _, ok := m.scrollOffset[i]; !ok {
			m.scrollOffset[i] = 0
		}
	}
}

func (m *BoardModel) SetFilter(fs FilterState) {
	m.filterState = fs
	m.applyFilter()
}

func (m *BoardModel) applyFilter() {
	// Clear column items
	for i := range m.columns {
		m.columns[i].Items = nil
	}

	// Build a map for quick column lookup
	colIndex := make(map[string]int)
	for i, col := range m.columns {
		colIndex[col.OptionID] = i
	}

	for _, item := range m.allItems {
		// Apply filters
		if !m.matchesFilter(item.Issue) {
			continue
		}

		idx, ok := colIndex[item.StatusID]
		if !ok {
			// Items without matching status go into first column
			if len(m.columns) > 0 {
				idx = 0
			} else {
				continue
			}
		}
		m.columns[idx].Items = append(m.columns[idx].Items, item)
	}
}

func (m *BoardModel) matchesFilter(issue domain.Issue) bool {
	if len(m.filterState.Labels) > 0 {
		if !matchLabels(issue, m.filterState.Labels) {
			return false
		}
	}
	if m.filterState.Milestone != "" {
		if issue.Milestone == nil || issue.Milestone.Title != m.filterState.Milestone {
			return false
		}
	}
	if len(m.filterState.Assignees) > 0 {
		if !matchAssignees(issue, m.filterState.Assignees) {
			return false
		}
	}
	return true
}

func matchLabels(issue domain.Issue, labels []string) bool {
	for _, fl := range labels {
		for _, il := range issue.Labels {
			if il.Name == fl {
				return true
			}
		}
	}
	return false
}

func matchAssignees(issue domain.Issue, assignees []string) bool {
	for _, fa := range assignees {
		for _, ia := range issue.Assignees {
			if ia.Login == fa {
				return true
			}
		}
	}
	return false
}

// columnItems returns the items in the given column index.
func (m BoardModel) columnItems(col int) []domain.ProjectItem {
	if col < 0 || col >= len(m.columns) {
		return nil
	}
	return m.columns[col].Items
}

// SelectedItem returns the currently focused ProjectItem, if any.
func (m BoardModel) SelectedItem() *domain.ProjectItem {
	items := m.columnItems(m.activeCol)
	idx := m.cursorIndex[m.activeCol]
	if idx < 0 || idx >= len(items) {
		return nil
	}
	item := items[idx]
	return &item
}

func (m BoardModel) Update(msg tea.Msg) (BoardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		items := m.columnItems(m.activeCol)
		switch {
		case msg.Code == 'h' || msg.Code == tea.KeyLeft:
			// Navigate to the previous visible column
			for i := m.activeCol - 1; i >= 0; i-- {
				if !m.hiddenCols[i] {
					m.activeCol = i
					break
				}
			}
		case msg.Code == 'l' || msg.Code == tea.KeyRight:
			// Navigate to the next visible column
			for i := m.activeCol + 1; i < len(m.columns); i++ {
				if !m.hiddenCols[i] {
					m.activeCol = i
					break
				}
			}
		case msg.Code == 'j' || msg.Code == tea.KeyDown:
			if m.cursorIndex[m.activeCol] < len(items)-1 {
				m.cursorIndex[m.activeCol]++
				m.adjustScroll()
			}
		case msg.Code == 'k' || msg.Code == tea.KeyUp:
			if m.cursorIndex[m.activeCol] > 0 {
				m.cursorIndex[m.activeCol]--
				m.adjustScroll()
			}
		case msg.Code == tea.KeyEnter:
			if len(items) > 0 {
				idx := m.cursorIndex[m.activeCol]
				if idx < len(items) {
					selected := items[idx].Issue
					m.selectedIssue = &selected
				}
			}
		case msg.Code == 'H':
			// Move status left
			m.wantStatusMove = -1
		case msg.Code == 'L':
			// Move status right
			m.wantStatusMove = 1
		case msg.Code == 'd':
			// Hide current column
			vis := m.visibleColumns()
			if len(vis) <= 1 {
				m.wantStatusMsg = "Cannot hide the last visible column"
			} else {
				m.hiddenCols[m.activeCol] = true
				// Move cursor to adjacent visible column
				moved := false
				// Try next column first
				for i := m.activeCol + 1; i < len(m.columns); i++ {
					if !m.hiddenCols[i] {
						m.activeCol = i
						moved = true
						break
					}
				}
				if !moved {
					// Try previous column
					for i := m.activeCol - 1; i >= 0; i-- {
						if !m.hiddenCols[i] {
							m.activeCol = i
							break
						}
					}
				}
			}
		case msg.Code == 'D':
			// Show all hidden columns
			m.hiddenCols = make(map[int]bool)
		}
	}
	return m, nil
}

func (m *BoardModel) adjustScroll() {
	maxVisible := m.maxVisibleCards()
	if maxVisible <= 0 {
		return
	}
	cursor := m.cursorIndex[m.activeCol]
	offset := m.scrollOffset[m.activeCol]
	if cursor < offset {
		m.scrollOffset[m.activeCol] = cursor
	}
	if cursor >= offset+maxVisible {
		m.scrollOffset[m.activeCol] = cursor - maxVisible + 1
	}
}

func (m BoardModel) maxVisibleCards() int {
	cardHeight := 4
	available := m.height - 3
	if available <= 0 {
		return 5
	}
	return available / cardHeight
}

func (m BoardModel) loadIssues(svc *service.IssueService) tea.Cmd {
	if svc == nil {
		return nil
	}
	return func() tea.Msg {
		issues, _, err := svc.ListIssues(context.Background(), service.ListIssuesOptions{
			PerPage: 100,
		})
		if err != nil {
			return issuesLoadedMsg{err: err}
		}
		return issuesLoadedMsg{issues: issues}
	}
}

func (m BoardModel) loadProjectData(projectSvc *service.ProjectService, projectNumber int) tea.Cmd {
	if projectSvc == nil {
		return nil
	}
	return func() tea.Msg {
		info, err := projectSvc.GetProjectFields(context.Background(), projectNumber)
		if err != nil {
			return projectDataMsg{err: err}
		}

		var allItems []domain.ProjectItem
		cursor := ""
		for {
			items, pageInfo, err := projectSvc.GetProjectItems(context.Background(), info.ID, cursor)
			if err != nil {
				return projectDataMsg{err: err}
			}
			allItems = append(allItems, items...)
			if !pageInfo.HasNextPage {
				break
			}
			cursor = pageInfo.EndCursor
		}

		return projectDataMsg{info: info, items: allItems}
	}
}

type projectDataMsg struct {
	info  domain.ProjectInfo
	items []domain.ProjectItem
	err   error
}

type statusMoveMsg struct {
	err error
}

func (m BoardModel) View() string {
	if m.loading {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Loading issues...")
	}

	visCols := m.visibleColumns()
	numCols := len(visCols)
	if numCols == 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "No columns to display")
	}

	colWidth := m.width/numCols - 2
	if colWidth < 20 {
		colWidth = 20
	}

	var cols []string
	for _, i := range visCols {
		cols = append(cols, m.renderColumn(m.columns[i].Name, i, colWidth))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	// Filter bar
	if m.hasActiveFilter() {
		filterBar := dimStyle.Render("Filter: " + m.filterDescription())
		board = lipgloss.JoinVertical(lipgloss.Left, filterBar, board)
	}

	return board
}

func (m BoardModel) renderColumn(title string, colIdx int, width int) string {
	items := m.columnItems(colIdx)
	isActive := m.activeCol == colIdx

	titleStr := columnTitleStyle.Render(fmt.Sprintf("%s (%d)", title, len(items)))
	if isActive {
		titleStr = columnTitleStyle.Foreground(lipgloss.Color("170")).Render(fmt.Sprintf("▸ %s (%d)", title, len(items)))
	}

	if len(items) == 0 {
		empty := dimStyle.Render("No issues")
		return columnStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, titleStr, empty))
	}

	maxVisible := m.maxVisibleCards()
	offset := m.scrollOffset[colIdx]
	end := offset + maxVisible
	if end > len(items) {
		end = len(items)
	}
	visible := items[offset:end]

	var cards []string
	for i, item := range visible {
		globalIdx := offset + i
		selected := isActive && globalIdx == m.cursorIndex[colIdx]
		cards = append(cards, renderCard(item.Issue, selected, width-4))
	}

	if offset > 0 {
		cards = append([]string{dimStyle.Render("  ▲ more")}, cards...)
	}
	if end < len(items) {
		cards = append(cards, dimStyle.Render("  ▼ more"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, cards...)
	return columnStyle.Width(width).Render(lipgloss.JoinVertical(lipgloss.Left, titleStr, content))
}

func renderCard(issue domain.Issue, selected bool, width int) string {
	style := cardStyle.Width(width)
	if selected {
		style = selectedCardStyle.Width(width)
	}

	number := dimStyle.Render(fmt.Sprintf("#%d", issue.Number))
	title := issue.Title
	if len(title) > width-8 {
		title = title[:width-11] + "..."
	}

	var labels []string
	for _, l := range issue.Labels {
		color := lipgloss.Color("#" + l.Color)
		labels = append(labels, labelStyle.Background(color).Foreground(lipgloss.Color("0")).Render(l.Name))
	}

	var assignees []string
	for _, a := range issue.Assignees {
		assignees = append(assignees, "@"+a.Login)
	}

	line1 := fmt.Sprintf("%s %s", number, title)
	var lines []string
	lines = append(lines, line1)
	if len(labels) > 0 {
		lines = append(lines, strings.Join(labels, " "))
	}
	if len(assignees) > 0 {
		lines = append(lines, dimStyle.Render(strings.Join(assignees, " ")))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m BoardModel) hasActiveFilter() bool {
	return len(m.filterState.Labels) > 0 || len(m.filterState.Assignees) > 0 || m.filterState.Milestone != ""
}

func (m BoardModel) filterDescription() string {
	var parts []string
	if len(m.filterState.Labels) > 0 {
		parts = append(parts, "labels:"+strings.Join(m.filterState.Labels, ","))
	}
	if len(m.filterState.Assignees) > 0 {
		parts = append(parts, "assignees:"+strings.Join(m.filterState.Assignees, ","))
	}
	if m.filterState.Milestone != "" {
		parts = append(parts, "milestone:"+m.filterState.Milestone)
	}
	return strings.Join(parts, " ")
}
