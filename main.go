package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
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
	)

	pflag.StringVarP(&repoFlag, "repo", "R", "", "Repository in owner/name format")
	pflag.IntVarP(&projectNumber, "project", "p", 0, "GitHub Project number to use for kanban columns")
	pflag.BoolVar(&showVersion, "version", false, "Show version")
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

	var projectSvc *service.ProjectService
	if projectNumber > 0 {
		projectSvc = service.NewProjectService(client, repoInfo.Owner, repoInfo.Name)
	}

	app := ui.NewAppModel(issueSvc, repoSvc, projectSvc, projectNumber)

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
