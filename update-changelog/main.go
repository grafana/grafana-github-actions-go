package main

import (
	"context"
	"fmt"
	"grafana-github-actions-go/internal/changelog"
	"grafana-github-actions-go/internal/git"
	"grafana-github-actions-go/internal/toolkit"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func main() {
	var version string
	var changelogFile string
	var repository string
	var repositoryPath string
	var ref string
	var targetBranch string
	var preview bool
	pflag.BoolVar(&preview, "preview", false, "Render a preview of the changelog entry without updating any files")
	pflag.StringVar(&version, "version", "", "Version to target")
	pflag.StringVar(&changelogFile, "changelog-file", "", "Path to changelog file to inject the new entry into")
	pflag.StringVar(&repository, "repo", os.Getenv("GITHUB_REPOSITORY"), "GitHub repository to clone and update")
	pflag.StringVar(&repositoryPath, "repo-path", "", "Path to an already check out version of repo")
	pflag.StringVar(&ref, "ref", os.Getenv("GITHUB_REF_NAME"), "Git branch to update the changelog in")
	pflag.StringVar(&targetBranch, "target-branch", "update-changelog", "Name of the branch to use for the pull-request")
	pflag.Parse()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())
	tk, err := toolkit.Init(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize toolkit")
	}

	if version == "" {
		version = tk.GetInput("version", nil)
	}
	if version == "" {
		logger.Fatal().Msg("No version specified")
	}

	body, err := changelog.Build(ctx, version, tk)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to build changelog")
	}

	if preview {
		fmt.Println(body.ToMarkdown())
		return
	}

	if changelogFile != "" {
		input, err := os.Open(changelogFile)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to open changelog file")
		}
		defer input.Close()

		if err := changelog.UpdateFile(ctx, os.Stdout, input, body); err != nil {
			logger.Fatal().Err(err).Msg("Failed to update changelog file")
		}
	} else if repository != "" {
		// If a changelog repository is provided, clone that repo at the
		// provided revision and use the changelog from there.
		elems := strings.Split(repository, "/")
		repoOwner := elems[0]
		repoRepo := elems[1]
		logger = logger.With().Str("repo", repoRepo).Str("owner", repoOwner).Str("targetBranch", targetBranch).Logger()
		branchExists, err := tk.BranchExists(ctx, repoOwner, repoRepo, targetBranch)
		title := fmt.Sprintf("Changelog: Updated changelog for %s", version)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to check if branch exists")
		}
		if branchExists {
			logger.Info().Msg("Target branch already exists. Pending PRs will be closed and the branch rebuilt.")
		}

		// Operate inside a temporary folder
		if repositoryPath == "" {
			tmpDir, err := os.MkdirTemp("", "update-changelog")
			defer os.RemoveAll(tmpDir)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to create a temporary directory for creating a checkout in")
			}
			if err := tk.CloneRepository(ctx, tmpDir, repository); err != nil {
				logger.Fatal().Err(err).Msg("Failed to clone repository")
			}
			repositoryPath = tmpDir
		}

		gitRepo := git.NewRepository(repositoryPath)

		if err := gitRepo.Exec(ctx, "switch", "--discard-changes", ref); err != nil {
			logger.Fatal().Err(err).Msg("Failed to switch to ref branch")
		}

		if err := gitRepo.Exec(ctx, "switch", "-C", targetBranch); err != nil {
			logger.Fatal().Err(err).Msg("Failed to switch to target branch")
		}

		if err := changelog.UpdateFileAtPath(ctx, filepath.Join(repositoryPath, "CHANGELOG.md"), body); err != nil {
			logger.Fatal().Err(err).Msg("Failed to update changelog")
		}

		if err := gitRepo.Exec(ctx, "add", "CHANGELOG.md"); err != nil {
			logger.Fatal().Err(err).Msg("Failed to add CHANGELOG.md")
		}

		if err := gitRepo.Exec(ctx, "commit", "-m", title); err != nil {
			logger.Fatal().Err(err).Msg("Failed to make commit")
		}

		if branchExists {
			logger.Info().Msg("Checking for existing pull requests")
			ghc := tk.GitHubClient()
			listOpts := github.PullRequestListOptions{}
			listOpts.Head = fmt.Sprintf("grafana:%s", targetBranch)
			listOpts.State = "open"
			pulls, _, err := ghc.PullRequests.List(ctx, repoOwner, repoRepo, &listOpts)
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to retrieve open pull-requests")
			}
			for _, pull := range pulls {
				{
					logger := logger.With().Str("pr", pull.GetTitle()).Logger()
					logger.Info().Msg("Closing PR")
					commentBody := "This pull request has been closed because an updated changelog and release notes have been generated."
					comment := github.IssueComment{}
					comment.Body = &commentBody

					if _, _, err := ghc.Issues.CreateComment(ctx, repoOwner, repoRepo, pull.GetNumber(), &comment); err != nil {
						logger.Fatal().Err(err).Msg("Failed to comment on pull-request")
					}
					closed := "closed"
					pull.State = &closed
					if _, _, err := ghc.PullRequests.Edit(ctx, repoOwner, repoRepo, pull.GetNumber(), pull); err != nil {
						logger.Fatal().Err(err).Msg("Failed to close pull-request")
					}
				}
			}
			if err := gitRepo.Exec(ctx, "push", "origin", "--delete", targetBranch); err != nil {
				logger.Fatal().Err(err).Msg("Failed to delete remote branch")
			}
		}

		if err := gitRepo.Exec(ctx, "push", "origin", targetBranch); err != nil {
			logger.Fatal().Err(err).Msg("Failed to push target branch")
		}

		isDraft := true
		pr := github.NewPullRequest{}
		pr.Title = &title
		pr.Draft = &isDraft
		pr.Base = &ref
		pr.Head = &targetBranch

		createPR, _, err := tk.GitHubClient().PullRequests.Create(ctx, repoOwner, repoRepo, &pr)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to create PR")
		}
		logger.Info().Msgf("New PR created at <%s>.", createPR.GetHTMLURL())
	}
}
