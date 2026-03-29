package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

// GitHubClient defines the interface for GitHub API access.
type GitHubClient interface {
	QueryGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error
	RESTGet(ctx context.Context, path string, result interface{}) error
	RESTPatch(ctx context.Context, path string, body interface{}, result interface{}) error
	RESTPost(ctx context.Context, path string, body interface{}, result interface{}) error
}

// Client implements GitHubClient using go-gh v2 or a raw HTTP client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewClient creates a Client using go-gh v2 authentication.
// It reads the token from gh auth and uses the default GitHub API base URL.
func NewClient() (*Client, error) {
	token, err := resolveToken()
	if err != nil {
		return nil, &domain.AppError{
			Code:    domain.ErrAuth,
			Message: "GitHub CLI authentication required. Run `gh auth login` to authenticate.",
			Err:     err,
		}
	}
	return &Client{
		httpClient: http.DefaultClient,
		baseURL:    "https://api.github.com",
		token:      token,
	}, nil
}

func (c *Client) doREST(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + "/" + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return classifyError(fmt.Errorf("request failed: %w", err))
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return classifyError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) RESTGet(ctx context.Context, path string, result interface{}) error {
	return c.doREST(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) RESTPost(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doREST(ctx, http.MethodPost, path, body, result)
}

func (c *Client) RESTPatch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doREST(ctx, http.MethodPatch, path, body, result)
}

func (c *Client) QueryGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	reqBody := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		reqBody["variables"] = variables
	}

	var graphQLResp struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := c.RESTPost(ctx, "graphql", reqBody, &graphQLResp); err != nil {
		return err
	}

	if len(graphQLResp.Errors) > 0 {
		msgs := make([]string, len(graphQLResp.Errors))
		for i, e := range graphQLResp.Errors {
			msgs[i] = e.Message
		}
		return &domain.AppError{
			Code:    domain.ErrValidation,
			Message: fmt.Sprintf("GraphQL error: %s", strings.Join(msgs, "; ")),
		}
	}

	if result != nil {
		if err := json.Unmarshal(graphQLResp.Data, result); err != nil {
			return fmt.Errorf("failed to decode GraphQL data: %w", err)
		}
	}
	return nil
}

// classifyError maps HTTP/API errors to domain error codes.
func classifyError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "401"):
		return &domain.AppError{
			Code:    domain.ErrAuth,
			Message: "Authentication failed. Run `gh auth login` to re-authenticate.",
			Err:     err,
		}
	case strings.Contains(msg, "403"):
		return &domain.AppError{
			Code:    domain.ErrPermission,
			Message: "Permission denied. Check your token scopes.",
			Err:     err,
		}
	case strings.Contains(msg, "404"):
		return &domain.AppError{
			Code:    domain.ErrNotFound,
			Message: "Resource not found.",
			Err:     err,
		}
	case strings.Contains(msg, "429"):
		return &domain.AppError{
			Code:    domain.ErrRateLimit,
			Message: "API rate limit exceeded. Please wait and try again.",
			Err:     err,
		}
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return &domain.AppError{
			Code:    domain.ErrNetwork,
			Message: "Request timed out. Check your network connection.",
			Err:     err,
		}
	default:
		return &domain.AppError{
			Code:    domain.ErrUnknown,
			Message: "An unexpected error occurred.",
			Err:     err,
		}
	}
}
