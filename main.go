package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/masamichhhhi/gh-tuissue/internal/cli"
	"github.com/masamichhhhi/gh-tuissue/internal/config"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
	"github.com/masamichhhhi/gh-tuissue/internal/repo"
	"github.com/masamichhhhi/gh-tuissue/internal/service"
	"github.com/masamichhhhi/gh-tuissue/internal/ui"
	"github.com/spf13/pflag"
)

var version = "dev"

func main() {
	var (
		repoFlag      string
		projectNumber int
		showVersion   bool
		showConfig    bool
	)

	pflag.StringVarP(&repoFlag, "repo", "R", "", "Repository in owner/name format")
	pflag.IntVarP(&projectNumber, "project", "p", 0, "GitHub Project number to use for kanban columns")
	pflag.BoolVar(&showVersion, "version", false, "Show version")
	pflag.BoolVar(&showConfig, "config", false, "Re-select GitHub Project binding")
	pflag.Parse()

	if showVersion {
		fmt.Printf("gh-tuissue version %s\n", version)
		os.Exit(0)
	}

	repoInfo, err := repo.Resolve(repoFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client, err := gh.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	issueSvc := service.NewIssueService(client, repoInfo.Owner, repoInfo.Name)
	repoSvc := service.NewRepoService(client, repoInfo.Owner, repoInfo.Name)
	projectSvc := service.NewProjectService(client, repoInfo.Owner, repoInfo.Name)

	// Resolve project number: --project flag > config file > CLI prompt
	if projectNumber == 0 && !showConfig {
		repoRoot := detectRepoRoot()
		if repoRoot != "" {
			cfg, _ := config.Load(repoRoot)
			if cfg != nil && cfg.ProjectNumber > 0 {
				projectNumber = cfg.ProjectNumber
			}
		}
	}

	// Show project selection prompt if needed
	if projectNumber == 0 || showConfig {
		repoRoot := detectRepoRoot()
		selected, err := cli.PromptProjectSelection(projectSvc, repoRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if selected > 0 {
			projectNumber = selected
		}
	}

	app := ui.NewAppModel(issueSvc, repoSvc, projectSvc, projectNumber)

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func detectRepoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
