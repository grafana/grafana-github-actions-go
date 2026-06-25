package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
)

// BackportTarget finds the most appropriate base branch (target) given a target name.
//
// For example: "v12.1.x" will return "release-12.1.15", assuming a branch
// named "release-12.1.15" is found among the candidates and its the most
// recent one.
func BackportTarget(version string, candidates []*github.Branch) (ghutil.Branch, error) {
	version = strings.TrimPrefix(version, "release-")
	labelString := strings.ReplaceAll(version, "x", "0")
	major, minor, _ := ghutil.MajorMinorPatch(labelString)

	return ghutil.MostRecentBranch(major, minor, candidates)
}

// BackportTargetsFromLabels extracts the list of target names from the list of
// labels of a PR. Typically, the prefix will be "backport ".
//
// For example ["backport v1", "backport r123"] will return ["v1", "r123"].
func BackportTargetsFromLabels(labels []string, prefix string) []string {
	var targets []string
	for _, label := range labels {
		if !strings.HasPrefix(label, prefix) {
			continue
		}

		n := strings.TrimSpace(strings.TrimPrefix(label, prefix))
		targets = append(targets, n)
	}

	return targets
}

// BackportTargets finds the most appropriate base branch (target) for each
// name.
//
// The accepted target names are:
//   - a branch name
//   - a short form `backport v11.2.x` and a release-branch form
//     `release-v2.8`, that finds the most recent `release-` branch that
//     matches.
func BackportTargets(
	ctx context.Context,
	log *slog.Logger,
	client ghutil.BranchClient,
	owner, repo string,
	names []string,
) ([]ghutil.Branch, error) {
	var (
		branches  []ghutil.Branch
		unmatched []string
	)

	// first pass matching target names as-is
	for _, n := range names {
		t, err := ghutil.GetReleaseBranchByName(ctx, client, owner, repo, n)
		if err != nil {
			unmatched = append(unmatched, n)
		} else {
			branches = append(branches, t)
		}
	}

	if len(unmatched) == 0 {
		return branches, nil
	}

	// second pass looking for branch matches
	candidates, err := ghutil.GetReleaseBranches(ctx, log, client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get release branches: %w", err)
	}

	for _, n := range unmatched {
		t, err := BackportTarget(n, candidates)
		if err != nil {
			return nil, fmt.Errorf("can't find a target for %s: %w", n, err)
		}
		branches = append(branches, t)
	}

	return branches, nil
}

func MergeBase(ctx context.Context, client *github.RepositoriesService, owner, repo, base, head string) (*github.Commit, error) {
	comp, _, err := client.CompareCommits(ctx, owner, repo, base, head, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return comp.MergeBaseCommit.Commit, nil
}
