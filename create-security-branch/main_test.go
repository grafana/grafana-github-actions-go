package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGitClient is a mock implementation of the GitClient interface
type MockGitClient struct {
	mock.Mock
}

func (m *MockGitClient) GetRef(ctx context.Context, owner string, repo string, ref string) (*github.Reference, *github.Response, error) {
	args := m.Called(ctx, owner, repo, ref)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*github.Response), args.Error(2)
	}
	return args.Get(0).(*github.Reference), args.Get(1).(*github.Response), args.Error(2)
}

func (m *MockGitClient) CreateRef(ctx context.Context, owner string, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	args := m.Called(ctx, owner, repo, ref)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*github.Response), args.Error(2)
	}
	return args.Get(0).(*github.Reference), args.Get(1).(*github.Response), args.Error(2)
}

func TestGetInputs(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		secNum      string
		ownerRepo   string
		expected    Inputs
		expectError bool
		errorMsg    string
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
			errorMsg:    "invalid repository format: invalid-repo, expected owner/repo",
		},
		{
			name:        "missing version",
			version:     "",
			secNum:      "01",
			ownerRepo:   "grafana/grafana-security-mirror",
			expected:    Inputs{},
			expectError: true,
			errorMsg:    "version is required",
		},
		{
			name:        "missing security branch number",
			version:     "12.0.1",
			secNum:      "",
			ownerRepo:   "grafana/grafana-security-mirror",
			expected:    Inputs{},
			expectError: true,
			errorMsg:    "security_branch_number is required",
		},
		{
			name:        "missing repository",
			version:     "12.0.1",
			secNum:      "01",
			ownerRepo:   "",
			expected:    Inputs{},
			expectError: true,
			errorMsg:    "repository is required",
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

			inputs, err := GetInputs()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Empty(t, inputs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, inputs)
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
				// Mock GetRef for checking existing branch
				m.On("GetRef", mock.Anything, "grafana", "grafana-security-mirror", "heads/12.0.1+security-01").
					Return(nil, &github.Response{}, errors.New("branch not found"))

				// Mock GetRef for getting base branch
				baseRef := &github.Reference{
					Ref: github.String("heads/release-12.0.1"),
					Object: &github.GitObject{
						SHA: github.String("abc123"),
					},
				}
				m.On("GetRef", mock.Anything, "grafana", "grafana-security-mirror", "heads/release-12.0.1").
					Return(baseRef, &github.Response{}, nil)

				// Mock CreateRef for creating new branch
				newRef := &github.Reference{
					Ref: github.String("heads/12.0.1+security-01"),
					Object: &github.GitObject{
						SHA: github.String("abc123"),
					},
				}
				m.On("CreateRef", mock.Anything, "grafana", "grafana-security-mirror", mock.Anything).
					Return(newRef, &github.Response{}, nil)
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
			errorMessage:   "invalid version format: invalid, expected x.y.z where x, y, and z are numbers",
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
				existingRef := &github.Reference{
					Ref: github.String("heads/12.0.1+security-01"),
				}
				m.On("GetRef", mock.Anything, "grafana", "grafana-security-mirror", "heads/12.0.1+security-01").
					Return(existingRef, &github.Response{}, nil)
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
				m.On("GetRef", mock.Anything, "grafana", "grafana-security-mirror", "heads/12.0.1+security-01").
					Return(nil, &github.Response{}, errors.New("branch not found"))
				m.On("GetRef", mock.Anything, "grafana", "grafana-security-mirror", "heads/release-12.0.1").
					Return(nil, &github.Response{}, errors.New("base branch not found"))
			},
			expectedBranch: "",
			expectError:    true,
			errorMessage:   "error getting base ref: base branch not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockGitClient)
			tt.mockSetup(mockClient)

			branch, err := CreateSecurityBranch(context.Background(), mockClient, tt.inputs)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Empty(t, branch)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBranch, branch)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
