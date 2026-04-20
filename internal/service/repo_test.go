package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepoService_ListLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/labels" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := []map[string]interface{}{
			{"name": "bug", "color": "d73a4a", "description": "Something broken"},
			{"name": "enhancement", "color": "a2eeef", "description": "New feature"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Encode failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewRepoService(client, "owner", "repo")

	labels, err := svc.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels failed: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
	if labels[0].Name != "bug" {
		t.Errorf("expected 'bug', got %q", labels[0].Name)
	}
	if labels[1].Color != "a2eeef" {
		t.Errorf("expected color 'a2eeef', got %q", labels[1].Color)
	}
}

func TestRepoService_ListCollaborators(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/collaborators" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := []map[string]interface{}{
			{"login": "user1", "name": "User One"},
			{"login": "user2", "name": "User Two"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Encode failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewRepoService(client, "owner", "repo")

	users, err := svc.ListCollaborators(context.Background())
	if err != nil {
		t.Fatalf("ListCollaborators failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Login != "user1" {
		t.Errorf("expected 'user1', got %q", users[0].Login)
	}
}

func TestRepoService_ListMilestones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/milestones" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := []map[string]interface{}{
			{"number": 1, "title": "v1.0", "state": "open", "due_on": nil},
			{"number": 2, "title": "v2.0", "state": "open", "due_on": "2026-06-01T00:00:00Z"},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("Encode failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewRepoService(client, "owner", "repo")

	milestones, err := svc.ListMilestones(context.Background())
	if err != nil {
		t.Fatalf("ListMilestones failed: %v", err)
	}
	if len(milestones) != 2 {
		t.Fatalf("expected 2 milestones, got %d", len(milestones))
	}
	if milestones[0].Title != "v1.0" {
		t.Errorf("expected 'v1.0', got %q", milestones[0].Title)
	}
	if milestones[0].DueOn != nil {
		t.Errorf("expected nil DueOn for v1.0")
	}
	if milestones[1].DueOn == nil {
		t.Fatal("expected non-nil DueOn for v2.0")
	}
}

func TestRepoService_ListLabels_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]interface{}{}); err != nil {
			t.Errorf("Encode failed: %v", err)
		}
	}))
	defer server.Close()

	client := newTestClient(server)
	svc := NewRepoService(client, "owner", "repo")

	labels, err := svc.ListLabels(context.Background())
	if err != nil {
		t.Fatalf("ListLabels failed: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(labels))
	}
}
