package domain

import (
	"fmt"
	"time"
)

type IssueState string

const (
	IssueOpen   IssueState = "OPEN"
	IssueClosed IssueState = "CLOSED"
)

type Issue struct {
	Number    int
	NodeID    string
	Title     string
	Body      string
	State     IssueState
	URL       string
	Author    User
	Labels    []Label
	Assignees []User
	Milestone *Milestone
	Comments  []Comment
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i Issue) IsOpen() bool {
	return i.State == IssueOpen
}

type Label struct {
	Name        string
	Color       string
	Description string
}

type User struct {
	Login string
	Name  string
}

type Milestone struct {
	Number int
	Title  string
	State  string
	DueOn  *time.Time
}

type Comment struct {
	ID        int
	Body      string
	Author    User
	CreatedAt time.Time
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

// Project V2 types

type ProjectSummary struct {
	ID     string
	Number int
	Title  string
}

type ProjectInfo struct {
	ID          string
	Title       string
	StatusField StatusField
}

type StatusField struct {
	ID      string
	Name    string
	Options []StatusOption
}

type StatusOption struct {
	ID   string
	Name string
}

type ProjectItem struct {
	ItemID   string
	Issue    Issue
	StatusID string
}

type ErrorCode int

const (
	ErrAuth              ErrorCode = iota
	ErrNetwork
	ErrNotFound
	ErrPermission
	ErrValidation
	ErrRateLimit
	ErrProjectNotFound
	ErrStatusFieldMissing
	ErrUnknown
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}
