package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// MockClient implements the `Client`` interface for testing purposes
type MockClient struct {
	EditPRCalled        bool
	CreateCommentCalled bool
	ShouldError         bool
	PrevBranch          string
	Comment             string
}

func (m *MockClient) EditPR(ctx context.Context, number int, branch string) error {
	m.EditPRCalled = true
	m.PrevBranch = branch
	if m.ShouldError {
		return fmt.Errorf("MOCK error: failed to edit PR %d to target %s", number, branch)
	}
	return nil
}

func (m *MockClient) CreateComment(ctx context.Context, number int, body string) error {
	m.CreateCommentCalled = true
	m.Comment = body
	if m.ShouldError {
		return fmt.Errorf("MOCK error: failed to create comment for PR %d", number)
	}
	return nil
}

func TestUpdatePRBranch(t *testing.T) {
	tests := []struct {
		name        string
		shouldError bool
		branch      string
	}{
		{
			name:        "successful update",
			shouldError: false,
			branch:      "new-branch",
		},
		{
			name:        "failed update",
			shouldError: true,
			branch:      "new-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{ShouldError: tt.shouldError}
			pr := PullRequestInfo{Number: 1, AuthorName: "testuser"}
			
			err := UpdateBaseBranch(context.Background(), mock, pr, tt.branch)
			
			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !mock.EditPRCalled {
				t.Error("EditPR was not called")
			}
			if !tt.shouldError && mock.PrevBranch != tt.branch {
				t.Errorf("wrong branch used, got: %s, want: %s", mock.PrevBranch, tt.branch)
			}
		})
	}
}

func TestNotifyUser(t *testing.T) {
	tests := []struct {
		name        string
		succeeded   bool
		shouldError bool
	}{
		{
			name:        "successful notification - success case",
			succeeded:   true,
			shouldError: false,
		},
		{
			name:        "successful notification - failure case",
			succeeded:   false,
			shouldError: false,
		},
		{
			name:        "failed notification",
			succeeded:   true,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockClient{ShouldError: tt.shouldError}
			pr := PullRequestInfo{Number: 1, AuthorName: "testuser"}
			
			err := NotifyUser(context.Background(), mock, pr, "old-branch", "new-branch", tt.succeeded)
			
			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !mock.CreateCommentCalled {
				t.Error("CreateComment was not called")
			}
			if !tt.shouldError {
				if !strings.Contains(mock.Comment, "@testuser") {
					t.Error("comment should mention user")
				}
				if !strings.Contains(mock.Comment, "old-branch") {
					t.Error("comment should mention old branch")
				}
				if !strings.Contains(mock.Comment, "new-branch") {
					t.Error("comment should mention new branch")
				}
			}
		})
	}
}

func TestBuildNotificationComment(t *testing.T) {
	tests := []struct {
		name      string
		author    string
		prev      string
		next      string
		succeeded bool
		want      []string // strings that should be in the comment
	}{
		{
			name:      "success message",
			author:    "testuser",
			prev:      "old-branch",
			next:      "new-branch",
			succeeded: true,
			want:      []string{"@testuser", "old-branch", "new-branch", "automatically updated"},
		},
		{
			name:      "failure message",
			author:    "testuser",
			prev:      "old-branch",
			next:      "new-branch",
			succeeded: false,
			want:      []string{"@testuser", "old-branch", "new-branch", "manually update"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := buildComment(tt.author, tt.prev, tt.next, tt.succeeded)
			
			for _, want := range tt.want {
				if !strings.Contains(comment, want) {
					t.Errorf("comment should contain %q", want)
				}
			}
		})
	}
}
