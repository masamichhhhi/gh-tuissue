package repo

import (
	"testing"
)

func TestParseRepoFlag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RepoInfo
		wantErr bool
	}{
		{
			name:  "valid owner/name",
			input: "owner/repo-name",
			want:  RepoInfo{Owner: "owner", Name: "repo-name", Host: "github.com"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no slash",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "too many slashes",
			input:   "a/b/c",
			wantErr: true,
		},
		{
			name:    "empty owner",
			input:   "/repo",
			wantErr: true,
		},
		{
			name:    "empty name",
			input:   "owner/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRepoFlag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRepoFlag(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRepoFlag(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseRepoFlag(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseGitRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    RepoInfo
		wantErr bool
	}{
		{
			name: "HTTPS URL",
			url:  "https://github.com/owner/repo.git",
			want: RepoInfo{Owner: "owner", Name: "repo", Host: "github.com"},
		},
		{
			name: "HTTPS URL without .git",
			url:  "https://github.com/owner/repo",
			want: RepoInfo{Owner: "owner", Name: "repo", Host: "github.com"},
		},
		{
			name: "SSH URL",
			url:  "git@github.com:owner/repo.git",
			want: RepoInfo{Owner: "owner", Name: "repo", Host: "github.com"},
		},
		{
			name: "SSH URL without .git",
			url:  "git@github.com:owner/repo",
			want: RepoInfo{Owner: "owner", Name: "repo", Host: "github.com"},
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitRemoteURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseGitRemoteURL(%q) expected error, got nil", tt.url)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseGitRemoteURL(%q) unexpected error: %v", tt.url, err)
			}
			if got != tt.want {
				t.Errorf("ParseGitRemoteURL(%q) = %+v, want %+v", tt.url, got, tt.want)
			}
		})
	}
}
