package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
)

type ProjectService struct {
	client gh.GitHubClient
	owner  string
	repo   string
}

func NewProjectService(client gh.GitHubClient, owner, repo string) *ProjectService {
	return &ProjectService{client: client, owner: owner, repo: repo}
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]domain.ProjectSummary, error) {
	query := `query($owner: String!, $repo: String!) {
		repository(owner: $owner, name: $repo) {
			projectsV2(first: 20) {
				nodes {
					id
					number
					title
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"owner": s.owner,
		"repo":  s.repo,
	}

	var result struct {
		Repository struct {
			ProjectsV2 struct {
				Nodes []struct {
					ID     string  `json:"id"`
					Number float64 `json:"number"`
					Title  string  `json:"title"`
				} `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"repository"`
	}

	if err := s.client.QueryGraphQL(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	var projects []domain.ProjectSummary
	for _, n := range result.Repository.ProjectsV2.Nodes {
		projects = append(projects, domain.ProjectSummary{
			ID:     n.ID,
			Number: int(n.Number),
			Title:  n.Title,
		})
	}
	return projects, nil
}

func (s *ProjectService) GetProjectFields(ctx context.Context, projectNumber int) (domain.ProjectInfo, error) {
	query := `query($owner: String!, $repo: String!, $projectNumber: Int!) {
		repository(owner: $owner, name: $repo) {
			projectV2(number: $projectNumber) {
				id
				title
				fields(first: 20) {
					nodes {
						__typename
						... on ProjectV2SingleSelectField {
							id
							name
							options {
								id
								name
							}
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"owner":         s.owner,
		"repo":          s.repo,
		"projectNumber": projectNumber,
	}

	var result struct {
		Repository struct {
			ProjectV2 struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Fields struct {
					Nodes []json.RawMessage `json:"nodes"`
				} `json:"fields"`
			} `json:"projectV2"`
		} `json:"repository"`
	}

	if err := s.client.QueryGraphQL(ctx, query, variables, &result); err != nil {
		return domain.ProjectInfo{}, err
	}

	pv2 := result.Repository.ProjectV2
	info := domain.ProjectInfo{
		ID:    pv2.ID,
		Title: pv2.Title,
	}

	for _, raw := range pv2.Fields.Nodes {
		var field struct {
			Typename string `json:"__typename"`
			ID       string `json:"id"`
			Name     string `json:"name"`
			Options  []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"options"`
		}
		if err := json.Unmarshal(raw, &field); err != nil {
			continue
		}
		if field.Typename == "ProjectV2SingleSelectField" && strings.EqualFold(field.Name, "Status") {
			info.StatusField = domain.StatusField{
				ID:   field.ID,
				Name: field.Name,
			}
			for _, opt := range field.Options {
				info.StatusField.Options = append(info.StatusField.Options, domain.StatusOption{
					ID:   opt.ID,
					Name: opt.Name,
				})
			}
			return info, nil
		}
	}

	return domain.ProjectInfo{}, &domain.AppError{
		Code:    domain.ErrStatusFieldMissing,
		Message: "Status field not found in project. Ensure the project has a single-select field named 'Status'.",
	}
}

func (s *ProjectService) GetProjectItems(ctx context.Context, projectID string, cursor string) ([]domain.ProjectItem, domain.PageInfo, error) {
	query := `query($projectId: ID!, $cursor: String) {
		node(id: $projectId) {
			... on ProjectV2 {
				items(first: 50, after: $cursor) {
					pageInfo { hasNextPage endCursor }
					nodes {
						id
						fieldValues(first: 20) {
							nodes {
								__typename
								... on ProjectV2ItemFieldSingleSelectValue {
									optionId
									name
									field { ... on ProjectV2SingleSelectField { name } }
								}
							}
						}
						content {
							__typename
							... on Issue {
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
						}
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"projectId": projectID,
	}
	if cursor != "" {
		variables["cursor"] = cursor
	}

	var result struct {
		Node struct {
			Items struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct {
					ID          string `json:"id"`
					FieldValues struct {
						Nodes []json.RawMessage `json:"nodes"`
					} `json:"fieldValues"`
					Content json.RawMessage `json:"content"`
				} `json:"nodes"`
			} `json:"items"`
		} `json:"node"`
	}

	if err := s.client.QueryGraphQL(ctx, query, variables, &result); err != nil {
		return nil, domain.PageInfo{}, err
	}

	pageInfo := domain.PageInfo{
		HasNextPage: result.Node.Items.PageInfo.HasNextPage,
		EndCursor:   result.Node.Items.PageInfo.EndCursor,
	}

	var items []domain.ProjectItem
	for _, node := range result.Node.Items.Nodes {
		item := domain.ProjectItem{
			ItemID: node.ID,
		}

		// Extract status from fieldValues
		for _, fvRaw := range node.FieldValues.Nodes {
			var fv struct {
				Typename string `json:"__typename"`
				OptionID string `json:"optionId"`
				Name     string `json:"name"`
				Field    struct {
					Name string `json:"name"`
				} `json:"field"`
			}
			if err := json.Unmarshal(fvRaw, &fv); err != nil {
				continue
			}
			if fv.Typename == "ProjectV2ItemFieldSingleSelectValue" && strings.EqualFold(fv.Field.Name, "Status") {
				item.StatusID = fv.OptionID
			}
		}

		// Parse content (Issue)
		var content struct {
			Typename  string  `json:"__typename"`
			ID        string  `json:"id"`
			Number    float64 `json:"number"`
			Title     string  `json:"title"`
			Body      string  `json:"body"`
			State     string  `json:"state"`
			URL       string  `json:"url"`
			CreatedAt string  `json:"createdAt"`
			UpdatedAt string  `json:"updatedAt"`
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
				Number float64 `json:"number"`
				Title  string  `json:"title"`
				State  string  `json:"state"`
				DueOn  string  `json:"dueOn"`
			} `json:"milestone"`
		}

		if err := json.Unmarshal(node.Content, &content); err != nil {
			continue
		}

		// Only process Issues (skip PRs and DraftIssues)
		if content.Typename != "" && content.Typename != "Issue" {
			continue
		}

		issue := domain.Issue{
			Number: int(content.Number),
			NodeID: content.ID,
			Title:  content.Title,
			Body:   content.Body,
			State:  domain.IssueState(content.State),
			URL:    content.URL,
			Author: domain.User{Login: content.Author.Login},
		}

		if t, err := time.Parse(time.RFC3339, content.CreatedAt); err == nil {
			issue.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339, content.UpdatedAt); err == nil {
			issue.UpdatedAt = t
		}

		for _, l := range content.Labels.Nodes {
			issue.Labels = append(issue.Labels, domain.Label{
				Name:        l.Name,
				Color:       l.Color,
				Description: l.Description,
			})
		}
		for _, a := range content.Assignees.Nodes {
			issue.Assignees = append(issue.Assignees, domain.User{
				Login: a.Login,
				Name:  a.Name,
			})
		}
		if content.Milestone != nil {
			ms := &domain.Milestone{
				Number: int(content.Milestone.Number),
				Title:  content.Milestone.Title,
				State:  content.Milestone.State,
			}
			if content.Milestone.DueOn != "" {
				if t, err := time.Parse(time.RFC3339, content.Milestone.DueOn); err == nil {
					ms.DueOn = &t
				}
			}
			issue.Milestone = ms
		}

		item.Issue = issue
		items = append(items, item)
	}

	return items, pageInfo, nil
}

func (s *ProjectService) MoveItemStatus(ctx context.Context, projectID, itemID, fieldID, optionID string) error {
	query := `mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
		updateProjectV2ItemFieldValue(
			input: { projectId: $projectId, itemId: $itemId, fieldId: $fieldId, value: { singleSelectOptionId: $optionId } }
		) {
			projectV2Item { id }
		}
	}`

	variables := map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"optionId":  optionID,
	}

	var result interface{}
	return s.client.QueryGraphQL(ctx, query, variables, &result)
}

func (s *ProjectService) AddItemToProject(ctx context.Context, projectID, contentID string) (string, error) {
	query := `mutation($projectId: ID!, $contentId: ID!) {
		addProjectV2ItemById(input: { projectId: $projectId, contentId: $contentId }) {
			item { id }
		}
	}`

	variables := map[string]interface{}{
		"projectId": projectID,
		"contentId": contentID,
	}

	var result struct {
		AddProjectV2ItemByID struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	}

	if err := s.client.QueryGraphQL(ctx, query, variables, &result); err != nil {
		return "", err
	}
	return result.AddProjectV2ItemByID.Item.ID, nil
}
