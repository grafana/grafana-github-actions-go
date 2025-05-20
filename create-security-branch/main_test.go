package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
)

// MockGitClient is a mock implementation of the GitClient interface
type MockGitClient struct {
	GetRefFunc    func(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error)
	CreateRefFunc func(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
}

func (m *MockGitClient) GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
	return m.GetRefFunc(ctx, owner, repo, ref)
}

func (m *MockGitClient) CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	return m.CreateRefFunc(ctx, owner, repo, ref)
}

func TestGetInputs(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		secNum      string
		ownerRepo   string
		expected    Inputs
		expectError bool
	}{
		{
			name:      "valid inputs",
			version:   "12.0.1",
			secNum:    "01",
			ownerRepo: "grafana/grafana-security-mirror",
			expected: Inputs{
				Version:           "12.0.1",
				SecurityBranchNum: "01",
				Owner:             "grafana",
				Repo:              "grafana-security-mirror",
			},
			expectError: false,
		},
		{
			name:        "invalid repository format",
			version:     "12.0.1",
			secNum:      "01",
			ownerRepo:   "invalid-repo",
			expected:    Inputs{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the test
			os.Setenv("INPUT_VERSION", tt.version)
			os.Setenv("INPUT_SECURITY_BRANCH_NUMBER", tt.secNum)
			os.Setenv("INPUT_REPOSITORY", tt.ownerRepo)
			defer func() {
				os.Unsetenv("INPUT_VERSION")
				os.Unsetenv("INPUT_SECURITY_BRANCH_NUMBER")
				os.Unsetenv("INPUT_REPOSITORY")
			}()

			if tt.expectError {
				require.Panics(t, func() { GetInputs() })
			} else {
				inputs := GetInputs()
				require.Equal(t, tt.expected, inputs)
			}
		})
	}
}

func TestCreateSecurityBranch(t *testing.T) {
	tests := []struct {
		name           string
		inputs         Inputs
		mockSetup      func(*MockGitClient)
		expectedBranch string
		expectError    bool
		errorMessage   string
	}{
		{
			name: "successful branch creation",
			inputs: Inputs{
				Version:           "12.0.1",
				SecurityBranchNum: "01",
				Owner:             "grafana",
				Repo:              "grafana-security-mirror",
			},
			mockSetup: func(m *MockGitClient) {
				m.GetRefFunc = func(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
					if ref == "heads/12.0.1+security-01" {
						return nil, &github.Response{}, errors.New("branch not found")
					}
					return &github.Reference{
						Ref: github.String("heads/release-12.0.1"),
						Object: &github.GitObject{
							SHA: github.String("abc123"),
						},
					}, &github.Response{}, nil
				}
				m.CreateRefFunc = func(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
					return ref, &github.Response{}, nil
				}
			},
			expectedBranch: "12.0.1+security-01",
			expectError:    false,
		},
		{
			name: "invalid version format",
			inputs: Inputs{
				Version:           "invalid",
				SecurityBranchNum: "01",
				Owner:             "grafana",
				Repo:              "grafana-security-mirror",
			},
			mockSetup:      func(m *MockGitClient) {},
			expectedBranch: "",
			expectError:    true,
			errorMessage:   "invalid version format: invalid, expected x.y.z",
		},
		{
			name: "invalid security branch number",
			inputs: Inputs{
				Version:           "12.0.1",
				SecurityBranchNum: "1",
				Owner:             "grafana",
				Repo:              "grafana-security-mirror",
			},
			mockSetup:      func(m *MockGitClient) {},
			expectedBranch: "",
			expectError:    true,
			errorMessage:   "invalid security branch number format: 1, expected two digits (e.g., 01)",
		},
		{
			name: "branch already exists",
			inputs: Inputs{
				Version:           "12.0.1",
				SecurityBranchNum: "01",
				Owner:             "grafana",
				Repo:              "grafana-security-mirror",
			},
			mockSetup: func(m *MockGitClient) {
				m.GetRefFunc = func(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
					return &github.Reference{
						Ref: github.String("heads/12.0.1+security-01"),
					}, &github.Response{}, nil
				}
			},
			expectedBranch: "",
			expectError:    true,
			errorMessage:   "security branch 12.0.1+security-01 already exists",
		},
		{
			name: "base branch not found",
			inputs: Inputs{
				Version:           "12.0.1",
				SecurityBranchNum: "01",
				Owner:             "grafana",
				Repo:              "grafana-security-mirror",
			},
			mockSetup: func(m *MockGitClient) {
				m.GetRefFunc = func(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
					if ref == "heads/12.0.1+security-01" {
						return nil, &github.Response{}, errors.New("branch not found")
					}
					return nil, &github.Response{}, errors.New("base branch not found")
				}
			},
			expectedBranch: "",
			expectError:    true,
			errorMessage:   "error getting base ref: base branch not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockGitClient{}
			tt.mockSetup(mockClient)

			branch, err := CreateSecurityBranch(context.Background(), mockClient, tt.inputs)

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, tt.errorMessage, err.Error())
				require.Empty(t, branch)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedBranch, branch)
			}
		})
	}
}
