package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
	gh "github.com/masamichhhhi/gh-tuissue/internal/github"
)

func newTestClient(server *httptest.Server) gh.GitHubClient {
	return gh.NewTestClient(server)
}

func TestIssueService_ListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		query, _ := reqBody["query"].(string)
		if !strings.Contains(query, "issues") {
			t.Errorf("expected GraphQL query to contain 'issues', got: %s", query)
		}

		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"repository": map[string]interface{}{
					"issues": map[string]interface{}{
						"nodes": []map[string]interface{}{
							{
								"number":    1,
								"title":     "First Issue",
								"body":      "Description",
								"state":     "OPEN",
								"url":       "https://github.com/owner/repo/issues/1",
								"createdAt": "2026-01-01T00:00:00Z",
								"updatedAt": "2026-01-02T00:00:00Z",
								"author":    map[string]interface{}{"login": "user1"},
								"labels": map[string]interface{}{
									"nodes": []map[string]interface{}{
										{"name": "bug", "color": "d73a4a", "description": "Something broken"},
									},
								},
								"assignees": map[string]interface{}{
									"nodes": []map[string]interface{}{
										{"login": "dev1", "name": "Developer 1"},
									},
								},
								"milestone": map[string]interface{}{
									"number": 1,
									"title":  "v1.0",
									"state":  "OPEN",
									"dueOn":  nil,
								},
							},
						},
						"pageInfo": map[string]interface{}{
							"hasNextPage": false,
							"endCursor":   "abc123",
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewIssueService(client, "owner", "repo")

	issues, pageInfo, err := svc.ListIssues(context.Background(), ListIssuesOptions{
		State:   domain.IssueOpen,
		PerPage: 20,
	})
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 1 {
		t.Errorf("expected issue number 1, got %d", issues[0].Number)
	}
	if issues[0].Title != "First Issue" {
		t.Errorf("expected title 'First Issue', got %q", issues[0].Title)
	}
	if issues[0].State != domain.IssueOpen {
		t.Errorf("expected state OPEN, got %q", issues[0].State)
	}
	if len(issues[0].Labels) != 1 || issues[0].Labels[0].Name != "bug" {
		t.Errorf("expected 1 label 'bug', got %+v", issues[0].Labels)
	}
	if len(issues[0].Assignees) != 1 || issues[0].Assignees[0].Login != "dev1" {
		t.Errorf("expected 1 assignee 'dev1', got %+v", issues[0].Assignees)
	}
	if issues[0].Milestone == nil || issues[0].Milestone.Title != "v1.0" {
		t.Errorf("expected milestone 'v1.0', got %+v", issues[0].Milestone)
	}
	if !pageInfo.HasNextPage {
		// This is fine — our test data says false
	}
	if pageInfo.EndCursor != "abc123" {
		t.Errorf("expected cursor 'abc123', got %q", pageInfo.EndCursor)
	}
}

func TestIssueService_CreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/graphql" {
			t.Error("CreateIssue should use REST, not GraphQL")
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "New Issue" {
			t.Errorf("expected title 'New Issue', got %v", body["title"])
		}

		resp := map[string]interface{}{
			"number":     42,
			"title":      "New Issue",
			"body":       "Body text",
			"state":      "open",
			"html_url":   "https://github.com/owner/repo/issues/42",
			"created_at": "2026-01-01T00:00:00Z",
			"updated_at": "2026-01-01T00:00:00Z",
			"user":       map[string]interface{}{"login": "creator"},
			"labels":     []interface{}{},
			"assignees":  []interface{}{},
			"milestone":  nil,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewIssueService(client, "owner", "repo")

	issue, err := svc.CreateIssue(context.Background(), CreateIssueInput{
		Title: "New Issue",
		Body:  "Body text",
	})
	if err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}
	if issue.Number != 42 {
		t.Errorf("expected issue number 42, got %d", issue.Number)
	}
	if issue.Title != "New Issue" {
		t.Errorf("expected title 'New Issue', got %q", issue.Title)
	}
}

func TestIssueService_UpdateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issues/1") {
			t.Errorf("expected path to contain /issues/1, got %s", r.URL.Path)
		}
		resp := map[string]interface{}{
			"number":     1,
			"title":      "Updated Title",
			"body":       "Updated body",
			"state":      "open",
			"html_url":   "https://github.com/owner/repo/issues/1",
			"created_at": "2026-01-01T00:00:00Z",
			"updated_at": "2026-01-02T00:00:00Z",
			"user":       map[string]interface{}{"login": "user1"},
			"labels":     []interface{}{},
			"assignees":  []interface{}{},
			"milestone":  nil,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewIssueService(client, "owner", "repo")

	title := "Updated Title"
	issue, err := svc.UpdateIssue(context.Background(), 1, UpdateIssueInput{
		Title: &title,
	})
	if err != nil {
		t.Fatalf("UpdateIssue failed: %v", err)
	}
	if issue.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", issue.Title)
	}
}

func TestIssueService_CloseAndReopenIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		state := body["state"].(string)
		resp := map[string]interface{}{
			"number":     1,
			"title":      "Issue",
			"body":       "",
			"state":      state,
			"html_url":   "https://github.com/owner/repo/issues/1",
			"created_at": "2026-01-01T00:00:00Z",
			"updated_at": "2026-01-02T00:00:00Z",
			"user":       map[string]interface{}{"login": "user1"},
			"labels":     []interface{}{},
			"assignees":  []interface{}{},
			"milestone":  nil,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewIssueService(client, "owner", "repo")

	closed, err := svc.CloseIssue(context.Background(), 1)
	if err != nil {
		t.Fatalf("CloseIssue failed: %v", err)
	}
	if closed.State != domain.IssueClosed {
		t.Errorf("expected CLOSED state, got %q", closed.State)
	}

	reopened, err := svc.ReopenIssue(context.Background(), 1)
	if err != nil {
		t.Fatalf("ReopenIssue failed: %v", err)
	}
	if reopened.State != domain.IssueOpen {
		t.Errorf("expected OPEN state, got %q", reopened.State)
	}
}

func TestIssueService_AddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		resp := map[string]interface{}{
			"id":         100,
			"body":       "A comment",
			"created_at": "2026-01-01T00:00:00Z",
			"user":       map[string]interface{}{"login": "commenter"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewIssueService(client, "owner", "repo")

	comment, err := svc.AddComment(context.Background(), 1, "A comment")
	if err != nil {
		t.Fatalf("AddComment failed: %v", err)
	}
	if comment.ID != 100 {
		t.Errorf("expected comment ID 100, got %d", comment.ID)
	}
	if comment.Body != "A comment" {
		t.Errorf("expected body 'A comment', got %q", comment.Body)
	}
}

func TestIssueService_ListComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]interface{}{
			{
				"id":         1,
				"body":       "Comment 1",
				"created_at": "2026-01-01T00:00:00Z",
				"user":       map[string]interface{}{"login": "user1"},
			},
			{
				"id":         2,
				"body":       "Comment 2",
				"created_at": "2026-01-02T00:00:00Z",
				"user":       map[string]interface{}{"login": "user2"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewIssueService(client, "owner", "repo")

	comments, err := svc.ListComments(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
	if comments[0].Body != "Comment 1" {
		t.Errorf("expected 'Comment 1', got %q", comments[0].Body)
	}
}
