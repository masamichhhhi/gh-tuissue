package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
)

type IssueService struct {
	client gh.GitHubClient
	owner  string
	repo   string
}

func NewIssueService(client gh.GitHubClient, owner, repo string) *IssueService {
	return &IssueService{client: client, owner: owner, repo: repo}
}

type ListIssuesOptions struct {
	State     domain.IssueState
	Labels    []string
	Assignee  string
	Milestone string
	Sort      string
	Direction string
	PerPage   int
	After     string
}

type CreateIssueInput struct {
	Title     string
	Body      string
	Labels    []string
	Assignees []string
	Milestone string
}

type UpdateIssueInput struct {
	Title     *string
	Body      *string
	Labels    *[]string
	Assignees *[]string
	Milestone *string
}

const listIssuesQuery = `query($owner: String!, $name: String!, $first: Int!, $after: String, $states: [IssueState!]) {
  repository(owner: $owner, name: $name) {
    issues(first: $first, after: $after, states: $states, orderBy: {field: UPDATED_AT, direction: DESC}) {
      nodes {
        id
        number
        title
        body
        state
        url
        createdAt
        updatedAt
        author { login }
        labels(first: 10) { nodes { name color description } }
        assignees(first: 10) { nodes { login name } }
        milestone { number title state dueOn }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
}`

type graphQLIssuesResponse struct {
	Repository struct {
		Issues struct {
			Nodes    []graphQLIssue `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"issues"`
	} `json:"repository"`
}

type graphQLIssue struct {
	ID        string `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels struct {
		Nodes []struct {
			Name        string `json:"name"`
			Color       string `json:"color"`
			Description string `json:"description"`
		} `json:"nodes"`
	} `json:"labels"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
			Name  string `json:"name"`
		} `json:"nodes"`
	} `json:"assignees"`
	Milestone *struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		DueOn  *string `json:"dueOn"`
	} `json:"milestone"`
}

func (s *IssueService) ListIssues(ctx context.Context, opts ListIssuesOptions) ([]domain.Issue, domain.PageInfo, error) {
	variables := map[string]interface{}{
		"owner": s.owner,
		"name":  s.repo,
		"first": opts.PerPage,
	}
	if opts.After != "" {
		variables["after"] = opts.After
	}
	if opts.State != "" {
		variables["states"] = []string{string(opts.State)}
	}

	var resp graphQLIssuesResponse
	if err := s.client.QueryGraphQL(ctx, listIssuesQuery, variables, &resp); err != nil {
		return nil, domain.PageInfo{}, err
	}

	issues := make([]domain.Issue, len(resp.Repository.Issues.Nodes))
	for i, node := range resp.Repository.Issues.Nodes {
		issues[i] = convertGraphQLIssue(node)
	}

	pageInfo := domain.PageInfo{
		HasNextPage: resp.Repository.Issues.PageInfo.HasNextPage,
		EndCursor:   resp.Repository.Issues.PageInfo.EndCursor,
	}

	return issues, pageInfo, nil
}

func (s *IssueService) GetIssue(ctx context.Context, number int) (domain.Issue, error) {
	var resp restIssueResponse
	path := fmt.Sprintf("repos/%s/%s/issues/%d", s.owner, s.repo, number)
	if err := s.client.RESTGet(ctx, path, &resp); err != nil {
		return domain.Issue{}, err
	}
	return convertRESTIssue(resp), nil
}

func (s *IssueService) CreateIssue(ctx context.Context, input CreateIssueInput) (domain.Issue, error) {
	body := map[string]interface{}{
		"title": input.Title,
		"body":  input.Body,
	}
	if len(input.Labels) > 0 {
		body["labels"] = input.Labels
	}
	if len(input.Assignees) > 0 {
		body["assignees"] = input.Assignees
	}
	if input.Milestone != "" {
		body["milestone"] = input.Milestone
	}

	var resp restIssueResponse
	path := fmt.Sprintf("repos/%s/%s/issues", s.owner, s.repo)
	if err := s.client.RESTPost(ctx, path, body, &resp); err != nil {
		return domain.Issue{}, err
	}
	return convertRESTIssue(resp), nil
}

func (s *IssueService) UpdateIssue(ctx context.Context, number int, input UpdateIssueInput) (domain.Issue, error) {
	body := map[string]interface{}{}
	if input.Title != nil {
		body["title"] = *input.Title
	}
	if input.Body != nil {
		body["body"] = *input.Body
	}
	if input.Labels != nil {
		body["labels"] = *input.Labels
	}
	if input.Assignees != nil {
		body["assignees"] = *input.Assignees
	}
	if input.Milestone != nil {
		body["milestone"] = *input.Milestone
	}

	var resp restIssueResponse
	path := fmt.Sprintf("repos/%s/%s/issues/%d", s.owner, s.repo, number)
	if err := s.client.RESTPatch(ctx, path, body, &resp); err != nil {
		return domain.Issue{}, err
	}
	return convertRESTIssue(resp), nil
}

func (s *IssueService) CloseIssue(ctx context.Context, number int) (domain.Issue, error) {
	body := map[string]interface{}{"state": "closed"}
	var resp restIssueResponse
	path := fmt.Sprintf("repos/%s/%s/issues/%d", s.owner, s.repo, number)
	if err := s.client.RESTPatch(ctx, path, body, &resp); err != nil {
		return domain.Issue{}, err
	}
	return convertRESTIssue(resp), nil
}

func (s *IssueService) ReopenIssue(ctx context.Context, number int) (domain.Issue, error) {
	body := map[string]interface{}{"state": "open"}
	var resp restIssueResponse
	path := fmt.Sprintf("repos/%s/%s/issues/%d", s.owner, s.repo, number)
	if err := s.client.RESTPatch(ctx, path, body, &resp); err != nil {
		return domain.Issue{}, err
	}
	return convertRESTIssue(resp), nil
}

func (s *IssueService) AddComment(ctx context.Context, number int, body string) (domain.Comment, error) {
	reqBody := map[string]interface{}{"body": body}
	var resp restCommentResponse
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", s.owner, s.repo, number)
	if err := s.client.RESTPost(ctx, path, reqBody, &resp); err != nil {
		return domain.Comment{}, err
	}
	return convertRESTComment(resp), nil
}

func (s *IssueService) ListComments(ctx context.Context, number int) ([]domain.Comment, error) {
	var resp []restCommentResponse
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", s.owner, s.repo, number)
	if err := s.client.RESTGet(ctx, path, &resp); err != nil {
		return nil, err
	}
	comments := make([]domain.Comment, len(resp))
	for i, c := range resp {
		comments[i] = convertRESTComment(c)
	}
	return comments, nil
}

// REST API response types

type restIssueResponse struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	} `json:"assignees"`
	Milestone *struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		DueOn  *string `json:"due_on"`
	} `json:"milestone"`
}

type restCommentResponse struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// Conversion helpers

func convertGraphQLIssue(node graphQLIssue) domain.Issue {
	issue := domain.Issue{
		Number: node.Number,
		NodeID: node.ID,
		Title:  node.Title,
		Body:   node.Body,
		State:  domain.IssueState(node.State),
		URL:    node.URL,
		Author: domain.User{Login: node.Author.Login},
	}

	issue.CreatedAt, _ = time.Parse(time.RFC3339, node.CreatedAt)
	issue.UpdatedAt, _ = time.Parse(time.RFC3339, node.UpdatedAt)

	for _, l := range node.Labels.Nodes {
		issue.Labels = append(issue.Labels, domain.Label{
			Name: l.Name, Color: l.Color, Description: l.Description,
		})
	}
	for _, a := range node.Assignees.Nodes {
		issue.Assignees = append(issue.Assignees, domain.User{Login: a.Login, Name: a.Name})
	}
	if node.Milestone != nil {
		ms := &domain.Milestone{
			Number: node.Milestone.Number,
			Title:  node.Milestone.Title,
			State:  node.Milestone.State,
		}
		if node.Milestone.DueOn != nil {
			t, err := time.Parse(time.RFC3339, *node.Milestone.DueOn)
			if err == nil {
				ms.DueOn = &t
			}
		}
		issue.Milestone = ms
	}
	return issue
}

func convertRESTIssue(resp restIssueResponse) domain.Issue {
	issue := domain.Issue{
		Number: resp.Number,
		Title:  resp.Title,
		Body:   resp.Body,
		State:  mapRESTState(resp.State),
		URL:    resp.HTMLURL,
		Author: domain.User{Login: resp.User.Login},
	}
	issue.CreatedAt, _ = time.Parse(time.RFC3339, resp.CreatedAt)
	issue.UpdatedAt, _ = time.Parse(time.RFC3339, resp.UpdatedAt)

	for _, l := range resp.Labels {
		issue.Labels = append(issue.Labels, domain.Label{
			Name: l.Name, Color: l.Color, Description: l.Description,
		})
	}
	for _, a := range resp.Assignees {
		issue.Assignees = append(issue.Assignees, domain.User{Login: a.Login, Name: a.Name})
	}
	if resp.Milestone != nil {
		ms := &domain.Milestone{
			Number: resp.Milestone.Number,
			Title:  resp.Milestone.Title,
			State:  resp.Milestone.State,
		}
		if resp.Milestone.DueOn != nil {
			t, err := time.Parse(time.RFC3339, *resp.Milestone.DueOn)
			if err == nil {
				ms.DueOn = &t
			}
		}
		issue.Milestone = ms
	}
	return issue
}

func convertRESTComment(resp restCommentResponse) domain.Comment {
	c := domain.Comment{
		ID:     resp.ID,
		Body:   resp.Body,
		Author: domain.User{Login: resp.User.Login},
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, resp.CreatedAt)
	return c
}

func mapRESTState(state string) domain.IssueState {
	switch strings.ToLower(state) {
	case "open":
		return domain.IssueOpen
	case "closed":
		return domain.IssueClosed
	default:
		return domain.IssueState(strings.ToUpper(state))
	}
}
