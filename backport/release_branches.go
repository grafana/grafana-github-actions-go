package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
)

func BackportTargets(branches []*github.Branch, labels []string) ([]ghutil.Branch, error) {
	targets := []ghutil.Branch{}
	for _, label := range labels {
		if !strings.HasPrefix(label, "backport ") {
			continue
		}

		target, err := BackportTarget(label, branches)
		if err != nil {
			return nil, fmt.Errorf("error getting target for backport label '%s': %w", label, err)
		}

		targets = append(targets, target)
	}

	return targets, nil
}

var (
	ErrorNotMerged = errors.New("pull request is not merged; nothing to do")
	ErrorBadAction = errors.New("unrecognized action")
	ErrorNoLabels  = errors.New("no labels found")
)

func BackportTargetsFromPayload(branches []*github.Branch, prInfo PrInfo) ([]ghutil.Branch, error) {
	if !prInfo.Pr.GetMerged() {
		return nil, ErrorNotMerged
	}

	if len(prInfo.Labels) == 0 {
		return nil, ErrorNoLabels
	}

	return BackportTargets(branches, prInfo.Labels)
}

// BackportTarget finds the most appropriate base branch (target) given the backport label 'label'
// This function takes the label, like `backport v11.2.x`, and finds the most recent `release-` branch
// that matches the pattern.
func BackportTarget(label string, branches []*github.Branch) (ghutil.Branch, error) {
	version := strings.TrimPrefix(label, "backport")
	labelString := strings.ReplaceAll(strings.TrimSpace(version), "x", "0")
	major, minor, _ := ghutil.MajorMinorPatch(labelString)

	return ghutil.MostRecentBranch(major, minor, branches)
}

func MergeBase(ctx context.Context, client *github.RepositoriesService, owner, repo, base, head string) (*github.Commit, error) {
	comp, _, err := client.CompareCommits(ctx, owner, repo, base, head, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return comp.MergeBaseCommit.Commit, nil
}
