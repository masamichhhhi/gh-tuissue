package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
)

func TestProjectService_ListProjects(t *testing.T) {
	mock := &gh.MockClient{
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			resp := map[string]interface{}{
				"repository": map[string]interface{}{
					"projectsV2": map[string]interface{}{
						"nodes": []interface{}{
							map[string]interface{}{
								"id":     "PVT_123",
								"number": float64(1),
								"title":  "Project Alpha",
							},
							map[string]interface{}{
								"id":     "PVT_456",
								"number": float64(2),
								"title":  "Project Beta",
							},
						},
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, result)
		},
	}

	svc := NewProjectService(mock, "owner", "repo")
	projects, err := svc.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("ListProjects() returned %d projects, want 2", len(projects))
	}
	if projects[0].Title != "Project Alpha" {
		t.Errorf("projects[0].Title = %q, want %q", projects[0].Title, "Project Alpha")
	}
	if projects[1].ID != "PVT_456" {
		t.Errorf("projects[1].ID = %q, want %q", projects[1].ID, "PVT_456")
	}
}

func TestProjectService_GetProjectFields(t *testing.T) {
	mock := &gh.MockClient{
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			resp := map[string]interface{}{
				"repository": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":    "PVT_123",
						"title": "My Project",
						"fields": map[string]interface{}{
							"nodes": []interface{}{
								map[string]interface{}{
									"__typename": "ProjectV2SingleSelectField",
									"id":         "PVTSSF_status",
									"name":       "Status",
									"options": []interface{}{
										map[string]interface{}{"id": "opt1", "name": "Todo"},
										map[string]interface{}{"id": "opt2", "name": "In Progress"},
										map[string]interface{}{"id": "opt3", "name": "Done"},
									},
								},
								map[string]interface{}{
									"__typename": "ProjectV2Field",
									"id":         "PVT_other",
									"name":       "Title",
								},
							},
						},
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, result)
		},
	}

	svc := NewProjectService(mock, "owner", "repo")
	info, err := svc.GetProjectFields(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetProjectFields() error = %v", err)
	}
	if info.ID != "PVT_123" {
		t.Errorf("ID = %q, want %q", info.ID, "PVT_123")
	}
	if info.StatusField.Name != "Status" {
		t.Errorf("StatusField.Name = %q, want %q", info.StatusField.Name, "Status")
	}
	if len(info.StatusField.Options) != 3 {
		t.Fatalf("StatusField.Options length = %d, want 3", len(info.StatusField.Options))
	}
	if info.StatusField.Options[1].Name != "In Progress" {
		t.Errorf("Options[1].Name = %q, want %q", info.StatusField.Options[1].Name, "In Progress")
	}
}

func TestProjectService_GetProjectFields_NoStatusField(t *testing.T) {
	mock := &gh.MockClient{
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			resp := map[string]interface{}{
				"repository": map[string]interface{}{
					"projectV2": map[string]interface{}{
						"id":    "PVT_123",
						"title": "My Project",
						"fields": map[string]interface{}{
							"nodes": []interface{}{
								map[string]interface{}{
									"__typename": "ProjectV2Field",
									"id":         "PVT_other",
									"name":       "Title",
								},
							},
						},
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, result)
		},
	}

	svc := NewProjectService(mock, "owner", "repo")
	_, err := svc.GetProjectFields(context.Background(), 1)
	if err == nil {
		t.Fatal("GetProjectFields() expected error for missing Status field")
	}
	appErr, ok := err.(*domain.AppError)
	if !ok {
		t.Fatalf("expected *domain.AppError, got %T", err)
	}
	if appErr.Code != domain.ErrStatusFieldMissing {
		t.Errorf("error code = %d, want ErrStatusFieldMissing (%d)", appErr.Code, domain.ErrStatusFieldMissing)
	}
}

func TestProjectService_GetProjectItems(t *testing.T) {
	mock := &gh.MockClient{
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			resp := map[string]interface{}{
				"node": map[string]interface{}{
					"items": map[string]interface{}{
						"pageInfo": map[string]interface{}{
							"hasNextPage": false,
							"endCursor":   "",
						},
						"nodes": []interface{}{
							map[string]interface{}{
								"id": "PVTI_1",
								"fieldValues": map[string]interface{}{
									"nodes": []interface{}{
										map[string]interface{}{
											"__typename": "ProjectV2ItemFieldSingleSelectValue",
											"optionId":   "opt1",
											"name":       "Todo",
											"field": map[string]interface{}{
												"name": "Status",
											},
										},
									},
								},
								"content": map[string]interface{}{
									"__typename": "Issue",
									"id":         "I_abc",
									"number":     float64(1),
									"title":      "First Issue",
									"body":       "body",
									"state":      "OPEN",
									"url":        "https://github.com/owner/repo/issues/1",
									"createdAt":  "2026-01-01T00:00:00Z",
									"updatedAt":  "2026-01-02T00:00:00Z",
									"author":     map[string]interface{}{"login": "dev"},
									"labels":     map[string]interface{}{"nodes": []interface{}{}},
									"assignees":  map[string]interface{}{"nodes": []interface{}{}},
									"milestone":  nil,
								},
							},
						},
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, result)
		},
	}

	svc := NewProjectService(mock, "owner", "repo")
	items, pageInfo, err := svc.GetProjectItems(context.Background(), "PVT_123", "")
	if err != nil {
		t.Fatalf("GetProjectItems() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("GetProjectItems() returned %d items, want 1", len(items))
	}
	if items[0].ItemID != "PVTI_1" {
		t.Errorf("ItemID = %q, want %q", items[0].ItemID, "PVTI_1")
	}
	if items[0].StatusID != "opt1" {
		t.Errorf("StatusID = %q, want %q", items[0].StatusID, "opt1")
	}
	if items[0].Issue.Number != 1 {
		t.Errorf("Issue.Number = %d, want 1", items[0].Issue.Number)
	}
	if items[0].Issue.NodeID != "I_abc" {
		t.Errorf("Issue.NodeID = %q, want %q", items[0].Issue.NodeID, "I_abc")
	}
	if pageInfo.HasNextPage {
		t.Error("expected HasNextPage = false")
	}
}

func TestProjectService_MoveItemStatus(t *testing.T) {
	called := false
	mock := &gh.MockClient{
		GraphQLHandler: func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
			called = true
			resp := map[string]interface{}{
				"updateProjectV2ItemFieldValue": map[string]interface{}{
					"projectV2Item": map[string]interface{}{
						"id": "PVTI_1",
					},
				},
			}
			data, _ := json.Marshal(resp)
			return json.Unmarshal(data, result)
		},
	}

	svc := NewProjectService(mock, "owner", "repo")
	err := svc.MoveItemStatus(context.Background(), "PVT_123", "PVTI_1", "PVTSSF_status", "opt2")
	if err != nil {
		t.Fatalf("MoveItemStatus() error = %v", err)
	}
	if !called {
		t.Error("expected GraphQL mutation to be called")
	}
}
