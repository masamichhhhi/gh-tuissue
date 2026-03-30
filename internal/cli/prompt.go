package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/masamichhhhi/gh-tuissue/internal/config"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	"github.com/masamichhhhi/gh-tuissue/internal/service"
)

// ProjectLister abstracts project listing for testability.
type ProjectLister interface {
	ListProjects(ctx context.Context) ([]domain.ProjectSummary, error)
}

// PromptProjectSelection displays CLI inline prompts for project binding.
// Returns the selected project number (0 means no project binding).
func PromptProjectSelection(projectSvc *service.ProjectService, repoRoot string) (int, error) {
	return promptProjectSelectionWith(projectSvc, repoRoot, runConfirm, runSelect)
}

// confirmFunc asks a yes/no question and returns the result.
type confirmFunc func(title string) (bool, error)

// selectFunc presents a list of options and returns the selected value.
type selectFunc func(title string, options []selectOption) (int, error)

type selectOption struct {
	Label string
	Value int
}

func promptProjectSelectionWith(lister ProjectLister, repoRoot string, confirm confirmFunc, sel selectFunc) (int, error) {
	bind, err := confirm("Bind a GitHub Project?")
	if err != nil {
		return 0, fmt.Errorf("prompt cancelled: %w", err)
	}

	if !bind {
		return 0, nil
	}

	fmt.Println("Fetching projects...")
	projects, err := lister.ListProjects(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to fetch projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found in this repository.")
		return 0, nil
	}

	options := make([]selectOption, len(projects))
	for i, p := range projects {
		options[i] = selectOption{
			Label: fmt.Sprintf("#%d %s", p.Number, p.Title),
			Value: p.Number,
		}
	}

	selected, err := sel("Select a project", options)
	if err != nil {
		return 0, fmt.Errorf("project selection cancelled: %w", err)
	}

	if repoRoot != "" {
		if err := config.Save(repoRoot, config.Config{ProjectNumber: selected}); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
		}
	}

	return selected, nil
}

// runConfirm is the default confirm implementation using huh.
func runConfirm(title string) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(title).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	return result, err
}

// runSelect is the default select implementation using huh.
func runSelect(title string, options []selectOption) (int, error) {
	huhOpts := make([]huh.Option[int], len(options))
	for i, o := range options {
		huhOpts[i] = huh.NewOption(o.Label, o.Value)
	}
	var selected int
	err := huh.NewSelect[int]().
		Title(title).
		Options(huhOpts...).
		Value(&selected).
		Run()
	return selected, err
}
