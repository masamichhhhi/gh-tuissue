package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type HelpModel struct{}

func NewHelpModel() HelpModel {
	return HelpModel{}
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	return m, nil
}

func (m HelpModel) View() string {
	sections := []struct {
		title string
		keys  [][]string
	}{
		{
			title: "Board",
			keys: [][]string{
				{"h/l / ←/→", "Switch column"},
				{"j/k / ↑/↓", "Move cursor up/down"},
				{"H / L", "Move issue status left/right"},
				{"Enter", "Open issue detail"},
				{"f", "Open filter panel"},
				{"r", "Refresh issues"},
				{"n", "Create new issue"},
			},
		},
		{
			title: "Detail",
			keys: [][]string{
				{"j/k / ↑/↓", "Scroll up/down"},
				{"s", "Toggle Open/Closed status"},
				{"e", "Edit body ($EDITOR)"},
				{"t", "Edit title"},
				{"l", "Edit labels"},
				{"a", "Edit assignees"},
				{"m", "Edit milestone"},
				{"c", "Add comment ($EDITOR)"},
				{"Esc", "Back to board"},
			},
		},
		{
			title: "Filter",
			keys: [][]string{
				{"j/k / ↑/↓", "Move cursor"},
				{"Tab", "Switch pane"},
				{"Space", "Toggle selection"},
				{"Enter", "Apply filter"},
				{"Esc", "Cancel"},
			},
		},
		{
			title: "Global",
			keys: [][]string{
				{"?", "Show this help"},
				{"q", "Quit"},
				{"Esc", "Go back"},
			},
		},
	}

	var content []string
	content = append(content, titleStyle.Render("Keyboard Shortcuts"))
	content = append(content, "")

	for _, section := range sections {
		content = append(content, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Render(section.title))
		for _, kv := range section.keys {
			key := lipgloss.NewStyle().Width(14).Bold(true).Render(kv[0])
			content = append(content, "  "+key+kv[1])
		}
		content = append(content, "")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(content, "\n"))
}
