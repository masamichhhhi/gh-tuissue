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

type projectSelectPhase int

const (
	PhaseConfirm projectSelectPhase = iota
	PhaseLoading
	PhaseSelectProject
)

// Messages

type projectsLoadedMsg struct {
	projects []domain.ProjectSummary
	err      error
}

type projectSelectedMsg struct {
	projectNumber int
}

type projectSkippedMsg struct{}

type ProjectSelectModel struct {
	phase      projectSelectPhase
	confirmIdx int // 0=Yes, 1=No
	projects   []domain.ProjectSummary
	cursor     int
	projectSvc *service.ProjectService
	repoRoot   string
	width      int
	height     int
	errMsg     string
}

func NewProjectSelectModel(projectSvc *service.ProjectService, repoRoot string) ProjectSelectModel {
	return ProjectSelectModel{
		phase:      PhaseConfirm,
		confirmIdx: 0,
		projectSvc: projectSvc,
		repoRoot:   repoRoot,
	}
}

func (m ProjectSelectModel) Init() tea.Cmd {
	return nil
}

func (m ProjectSelectModel) Update(msg tea.Msg) (ProjectSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch m.phase {
		case PhaseConfirm:
			return m.updateConfirm(msg)
		case PhaseSelectProject:
			return m.updateSelectProject(msg)
		}

	case projectsLoadedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("プロジェクト一覧の取得に失敗しました: %v", msg.err)
			m.phase = PhaseConfirm
			return m, nil
		}
		if len(msg.projects) == 0 {
			m.errMsg = "プロジェクトが見つかりませんでした"
			m.phase = PhaseConfirm
			return m, nil
		}
		m.projects = msg.projects
		m.cursor = 0
		m.phase = PhaseSelectProject
		return m, nil
	}

	return m, nil
}

func (m ProjectSelectModel) updateConfirm(msg tea.KeyPressMsg) (ProjectSelectModel, tea.Cmd) {
	switch {
	case msg.Code == 'j' || msg.Code == 'k' || msg.Code == tea.KeyDown || msg.Code == tea.KeyUp:
		if m.confirmIdx == 0 {
			m.confirmIdx = 1
		} else {
			m.confirmIdx = 0
		}
	case msg.Code == tea.KeyEnter:
		if m.confirmIdx == 0 {
			// Yes - load projects
			m.phase = PhaseLoading
			m.errMsg = ""
			return m, m.loadProjects()
		}
		// No - skip project selection
		return m, func() tea.Msg { return projectSkippedMsg{} }
	}
	return m, nil
}

func (m ProjectSelectModel) updateSelectProject(msg tea.KeyPressMsg) (ProjectSelectModel, tea.Cmd) {
	switch {
	case msg.Code == 'j' || msg.Code == tea.KeyDown:
		if m.cursor < len(m.projects)-1 {
			m.cursor++
		}
	case msg.Code == 'k' || msg.Code == tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case msg.Code == tea.KeyEnter:
		if m.cursor < len(m.projects) {
			selected := m.projects[m.cursor]
			// Save config
			if m.repoRoot != "" {
				_ = config.Save(m.repoRoot, config.Config{ProjectNumber: selected.Number})
			}
			number := selected.Number
			return m, func() tea.Msg { return projectSelectedMsg{projectNumber: number} }
		}
	case msg.Code == tea.KeyEscape:
		m.phase = PhaseConfirm
		return m, nil
	}
	return m, nil
}

func (m ProjectSelectModel) loadProjects() tea.Cmd {
	if m.projectSvc == nil {
		return func() tea.Msg {
			return projectsLoadedMsg{err: fmt.Errorf("project service not available")}
		}
	}
	svc := m.projectSvc
	return func() tea.Msg {
		projects, err := svc.ListProjects(context.Background())
		return projectsLoadedMsg{projects: projects, err: err}
	}
}

func (m ProjectSelectModel) View() string {
	var b strings.Builder

	switch m.phase {
	case PhaseConfirm:
		b.WriteString(titleStyle.Render("gh-tuissue セットアップ"))
		b.WriteString("\n\n")
		if m.errMsg != "" {
			b.WriteString(errorStyle.Render(m.errMsg))
			b.WriteString("\n\n")
		}
		b.WriteString("Projectを紐付けますか？\n\n")

		yes := "  Yes"
		no := "  No"
		if m.confirmIdx == 0 {
			yes = selectedCardStyle.Render("▸ Yes")
		} else {
			no = selectedCardStyle.Render("▸ No")
		}
		b.WriteString(yes + "\n")
		b.WriteString(no + "\n")

	case PhaseLoading:
		b.WriteString("プロジェクト一覧を取得中...")

	case PhaseSelectProject:
		b.WriteString(titleStyle.Render("プロジェクトを選択してください"))
		b.WriteString("\n\n")

		for i, p := range m.projects {
			prefix := "  "
			if i == m.cursor {
				prefix = "▸ "
			}
			line := fmt.Sprintf("%s#%d %s", prefix, p.Number, p.Title)
			if i == m.cursor {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(line)
			}
			b.WriteString(line + "\n")
		}
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m ProjectSelectModel) KeyHints() string {
	switch m.phase {
	case PhaseConfirm:
		return dimStyle.Render("j/k:選択 enter:決定")
	case PhaseSelectProject:
		return dimStyle.Render("j/k:選択 enter:決定 esc:戻る")
	default:
		return ""
	}
}
