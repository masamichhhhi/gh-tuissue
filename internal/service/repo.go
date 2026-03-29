package service

import (
	"context"
	"fmt"
	"time"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
)

type RepoService struct {
	client gh.GitHubClient
	owner  string
	repo   string
}

func NewRepoService(client gh.GitHubClient, owner, repo string) *RepoService {
	return &RepoService{client: client, owner: owner, repo: repo}
}

func (s *RepoService) ListLabels(ctx context.Context) ([]domain.Label, error) {
	var resp []struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}
	path := fmt.Sprintf("repos/%s/%s/labels", s.owner, s.repo)
	if err := s.client.RESTGet(ctx, path, &resp); err != nil {
		return nil, err
	}
	labels := make([]domain.Label, len(resp))
	for i, l := range resp {
		labels[i] = domain.Label{Name: l.Name, Color: l.Color, Description: l.Description}
	}
	return labels, nil
}

func (s *RepoService) ListCollaborators(ctx context.Context) ([]domain.User, error) {
	var resp []struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	path := fmt.Sprintf("repos/%s/%s/collaborators", s.owner, s.repo)
	if err := s.client.RESTGet(ctx, path, &resp); err != nil {
		return nil, err
	}
	users := make([]domain.User, len(resp))
	for i, u := range resp {
		users[i] = domain.User{Login: u.Login, Name: u.Name}
	}
	return users, nil
}

func (s *RepoService) ListMilestones(ctx context.Context) ([]domain.Milestone, error) {
	var resp []struct {
		Number int     `json:"number"`
		Title  string  `json:"title"`
		State  string  `json:"state"`
		DueOn  *string `json:"due_on"`
	}
	path := fmt.Sprintf("repos/%s/%s/milestones", s.owner, s.repo)
	if err := s.client.RESTGet(ctx, path, &resp); err != nil {
		return nil, err
	}
	milestones := make([]domain.Milestone, len(resp))
	for i, m := range resp {
		ms := domain.Milestone{Number: m.Number, Title: m.Title, State: m.State}
		if m.DueOn != nil {
			t, err := time.Parse(time.RFC3339, *m.DueOn)
			if err == nil {
				ms.DueOn = &t
			}
		}
		milestones[i] = ms
	}
	return milestones, nil
}
