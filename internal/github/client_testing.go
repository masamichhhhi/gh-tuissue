package github

import (
	"context"
	"net/http/httptest"
)

// newTestRESTClient creates a Client pointing at a test HTTP server (internal use).
func newTestRESTClient(server *httptest.Server) *Client {
	return &Client{
		httpClient: server.Client(),
		baseURL:    server.URL,
		token:      "test-token",
	}
}

// NewTestClient creates a Client pointing at a test HTTP server (exported for cross-package tests).
func NewTestClient(server *httptest.Server) *Client {
	return newTestRESTClient(server)
}

// MockClient implements GitHubClient for testing with handler functions.
type MockClient struct {
	GraphQLHandler func(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error
	RESTGetHandler func(ctx context.Context, path string, result interface{}) error
	RESTPatchHandler func(ctx context.Context, path string, body interface{}, result interface{}) error
	RESTPostHandler func(ctx context.Context, path string, body interface{}, result interface{}) error
}

func (m *MockClient) QueryGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	if m.GraphQLHandler != nil {
		return m.GraphQLHandler(ctx, query, variables, result)
	}
	return nil
}

func (m *MockClient) RESTGet(ctx context.Context, path string, result interface{}) error {
	if m.RESTGetHandler != nil {
		return m.RESTGetHandler(ctx, path, result)
	}
	return nil
}

func (m *MockClient) RESTPatch(ctx context.Context, path string, body interface{}, result interface{}) error {
	if m.RESTPatchHandler != nil {
		return m.RESTPatchHandler(ctx, path, body, result)
	}
	return nil
}

func (m *MockClient) RESTPost(ctx context.Context, path string, body interface{}, result interface{}) error {
	if m.RESTPostHandler != nil {
		return m.RESTPostHandler(ctx, path, body, result)
	}
	return nil
}
