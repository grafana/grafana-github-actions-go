package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghgql"
)

// RefClient is the subset of go-github's GitService that this package uses.
// Defining an interface keeps the code testable.
type RefClient interface {
	CreateRef(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error)
}

// SignedCommitClient is the subset of pkg/ghgql.Client that PublishViaAPI uses.
type SignedCommitClient interface {
	CreateSignedCommitOnBranch(ctx context.Context, repo, branch, expectedHeadOid, headline, body string, changes []ghgql.FileChange) (string, error)
}

// PublishViaAPI publishes the local cherry-picked commit to the destination repo via GitHub's
// GraphQL `createCommitOnBranch` mutation. The mutation creates a commit that is signed by
// GitHub's web-flow key (shown as Verified), unlike `git push` which produces an unsigned commit.
//
// The function assumes:
//   - The working directory is a checkout of opts.Owner/opts.Repository.
//   - HEAD is the cherry-picked commit (i.e. CreateCherryPickBranch has already run).
//   - origin/<opts.Target.Name> is reachable in the local refs.
//
// It (1) reads the cherry-picked commit's tree diff vs origin/<target> to build a FileChanges
// payload, (2) reads the original PR's author from opts.SourceSHA and appends a Co-authored-by
// trailer so attribution survives (the API-created commit's author/committer are the
// authenticated identity), (3) creates the head branch via the REST refs API, and (4) calls the
// createCommitOnBranch mutation to publish the commit.
func PublishViaAPI(ctx context.Context, runner CommandRunner, refClient RefClient, gqlClient SignedCommitClient, branch string, opts BackportOpts) error {
	target := opts.Target.Name

	baseSha, err := gitOutput(ctx, runner, "rev-parse", "origin/"+target)
	if err != nil {
		return fmt.Errorf("resolving origin/%s: %w", target, err)
	}

	changes, err := buildFileChanges(ctx, runner, baseSha, "HEAD")
	if err != nil {
		return fmt.Errorf("building file changes: %w", err)
	}
	if len(changes) == 0 {
		// The cherry-pick should always produce a non-empty diff against base; if not, something
		// went wrong upstream (e.g. the commit was already on the target branch). Fail explicitly
		// rather than letting the API return an opaque error.
		return fmt.Errorf("no file changes detected between origin/%s and HEAD after cherry-pick; refusing to create an empty commit", target)
	}

	commitMessage, err := gitOutput(ctx, runner, "log", "-1", "--format=%B", "HEAD")
	if err != nil {
		return fmt.Errorf("reading commit message: %w", err)
	}

	authorName, _ := gitOutput(ctx, runner, "log", "-1", "--format=%an", opts.SourceSHA)
	authorEmail, _ := gitOutput(ctx, runner, "log", "-1", "--format=%ae", opts.SourceSHA)
	commitMessage = appendCoAuthor(commitMessage, authorName, authorEmail)

	headline, body := splitMessage(commitMessage)

	// Create the head branch on GitHub pointing at the base SHA, so createCommitOnBranch has a
	// branch to extend. If the ref already exists (e.g. from a previous failed publish attempt
	// or a manual retry) and points at the same base SHA, treat that as success.
	if err := ensureRef(ctx, refClient, opts.Owner, opts.Repository, branch, baseSha); err != nil {
		return err
	}

	if _, err := gqlClient.CreateSignedCommitOnBranch(ctx, opts.Owner+"/"+opts.Repository, branch, baseSha, headline, body, changes); err != nil {
		return fmt.Errorf("creating signed commit on branch %s: %w", branch, err)
	}
	return nil
}

// ensureRef creates refs/heads/<branch> pointing at baseSha. If the ref already exists, the
// function succeeds silently — the next createCommitOnBranch call will use baseSha as
// expectedHeadOid and either advance the branch correctly or surface a mismatch.
func ensureRef(ctx context.Context, refClient RefClient, owner, repo, branch, baseSha string) error {
	_, _, err := refClient.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.String("refs/heads/" + branch),
		Object: &github.GitObject{SHA: github.String(baseSha)},
	})
	if err == nil {
		return nil
	}
	// go-github returns a *github.ErrorResponse for 4xx responses; the message includes
	// "Reference already exists" for the 422 case.
	if strings.Contains(err.Error(), "Reference already exists") {
		return nil
	}
	return fmt.Errorf("creating ref refs/heads/%s: %w", branch, err)
}

func gitOutput(ctx context.Context, runner CommandRunner, args ...string) (string, error) {
	out, err := runner.Run(ctx, "git", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// buildFileChanges parses `git diff --no-renames --name-status -z <from> <to>` and reads the
// added/modified files' contents from disk, returning a FileChange slice suitable for
// createCommitOnBranch. Renames are flattened into delete-plus-add via --no-renames.
func buildFileChanges(ctx context.Context, runner CommandRunner, from, to string) ([]ghgql.FileChange, error) {
	out, err := runner.Run(ctx, "git", "diff", "--no-renames", "--name-status", "-z", from, to)
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}

	// `-z` emits NUL-separated tokens alternating status and path. Without renames, every entry
	// is a (status, path) pair.
	tokens := strings.Split(out, "\x00")
	var changes []ghgql.FileChange
	for i := 0; i+1 < len(tokens); i += 2 {
		status := tokens[i]
		path := tokens[i+1]
		if status == "" || path == "" {
			continue
		}
		switch status[0] {
		case 'D':
			changes = append(changes, ghgql.FileChange{Path: path, Delete: true})
		case 'A', 'M', 'T':
			data, err := os.ReadFile(filepath.Clean(path))
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", path, err)
			}
			changes = append(changes, ghgql.FileChange{Path: path, Contents: data})
		}
	}
	return changes, nil
}

// appendCoAuthor returns the message with a Co-authored-by trailer naming the original commit's
// author. The trailer is omitted if name or email are empty.
func appendCoAuthor(message, name, email string) string {
	message = strings.TrimRight(message, "\n")
	if name == "" || email == "" {
		return message
	}
	return message + "\n\nCo-authored-by: " + name + " <" + email + ">"
}

// splitMessage splits a commit message into a headline (the first line) and a body (the rest),
// since createCommitOnBranch takes the two as separate inputs. Leading newlines on the body —
// the conventional blank line between subject and body — are trimmed because the API
// re-introduces the separator itself.
func splitMessage(message string) (headline, body string) {
	message = strings.TrimRight(message, "\n")
	parts := strings.SplitN(message, "\n", 2)
	headline = parts[0]
	if len(parts) > 1 {
		body = strings.TrimLeft(parts[1], "\n")
	}
	return headline, body
}
