package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/ghutil"
)

func BackportTargets(branches []string, labels []*github.Label) ([]string, error) {
	targets := []string{}
	for _, label := range labels {
		if !strings.HasPrefix(label.GetName(), "backport ") {
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
)

func BackportTargetsFromPayload(branches []string, payload *github.PullRequestTargetEvent) ([]string, error) {
	if !payload.PullRequest.GetMerged() {
		return nil, ErrorNotMerged
	}

	switch payload.GetAction() {
	case "labeled":
		return BackportTargets(branches, []*github.Label{payload.GetLabel()})
	case "closed":
		return BackportTargets(branches, payload.GetPullRequest().Labels)
	}

	return nil, ErrorBadAction
}

// BackportTarget finds the most appropriate base branch (target) given the backport label 'label'
// This function takes the label, like `backport v11.2.x`, and finds the most recent `release-` branch
// that matches the pattern.
func BackportTarget(label *github.Label, branches []string) (string, error) {
	version := strings.TrimPrefix(label.GetName(), "backport")
	labelString := strings.TrimSpace(version)
	major, minor, _ := ghutil.MajorMinorPatch(labelString)

	return ghutil.MostRecentBranch(major, minor, branches)
}
