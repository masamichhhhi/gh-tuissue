package github

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

func TestNewClient_interface(t *testing.T) {
	// Verify Client implements GitHubClient interface
	var _ GitHubClient = (*Client)(nil)
}

func TestClient_RESTGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/repos/owner/repo/issues" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"number": 1, "title": "issue1"},
		})
	}))
	defer server.Close()

	client := newTestRESTClient(server)
	var result []map[string]interface{}
	err := client.RESTGet(context.Background(), "repos/owner/repo/issues", &result)
	if err != nil {
		t.Fatalf("RESTGet failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
}

func TestClient_RESTPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "new issue" {
			t.Errorf("unexpected body: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"number": 42})
	}))
	defer server.Close()

	client := newTestRESTClient(server)
	body := map[string]interface{}{"title": "new issue"}
	var result map[string]interface{}
	err := client.RESTPost(context.Background(), "repos/owner/repo/issues", body, &result)
	if err != nil {
		t.Fatalf("RESTPost failed: %v", err)
	}
	if result["number"] != float64(42) {
		t.Errorf("expected number 42, got %v", result["number"])
	}
}

func TestClient_RESTPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"state": "closed"})
	}))
	defer server.Close()

	client := newTestRESTClient(server)
	body := map[string]interface{}{"state": "closed"}
	var result map[string]interface{}
	err := client.RESTPatch(context.Background(), "repos/owner/repo/issues/1", body, &result)
	if err != nil {
		t.Fatalf("RESTPatch failed: %v", err)
	}
}

func TestClient_RESTGet_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer server.Close()

	client := newTestRESTClient(server)
	var result interface{}
	err := client.RESTGet(context.Background(), "repos/owner/repo/issues", &result)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestClassifyError_Auth(t *testing.T) {
	err := classifyError(errors.New("HTTP 401: Bad credentials"))
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}
	if appErr.Code != domain.ErrAuth {
		t.Errorf("expected ErrAuth, got %d", appErr.Code)
	}
}

func TestClassifyError_NotFound(t *testing.T) {
	err := classifyError(errors.New("HTTP 404: Not Found"))
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}
	if appErr.Code != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %d", appErr.Code)
	}
}

func TestClassifyError_RateLimit(t *testing.T) {
	err := classifyError(errors.New("HTTP 429: rate limit exceeded"))
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}
	if appErr.Code != domain.ErrRateLimit {
		t.Errorf("expected ErrRateLimit, got %d", appErr.Code)
	}
}

func TestClassifyError_Forbidden(t *testing.T) {
	err := classifyError(errors.New("HTTP 403: Forbidden"))
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		t.Fatal("expected AppError")
	}
	if appErr.Code != domain.ErrPermission {
		t.Errorf("expected ErrPermission, got %d", appErr.Code)
	}
}
