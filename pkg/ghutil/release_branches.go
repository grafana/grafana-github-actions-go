package ghutil

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/grafana/grafana-github-actions-go/pkg/versions"
)

type BranchClient interface {
	ListBranches(ctx context.Context, owner string, repo string, opts *github.BranchListOptions) ([]*github.Branch, *github.Response, error)
	GetBranch(ctx context.Context, owner, repo, branch string, followRedirects bool) (*github.Branch, *github.Response, error)
}

type Branch struct {
	Name  string
	SHA   string
	Major string
	Minor string
	Patch string
}

func SortBranches(a, b Branch) int {
	aMajor, _ := strconv.Atoi(a.Major)
	aMinor, _ := strconv.Atoi(a.Minor)
	aPatch, _ := strconv.Atoi(a.Patch)

	bMajor, _ := strconv.Atoi(b.Major)
	bMinor, _ := strconv.Atoi(b.Minor)
	bPatch, _ := strconv.Atoi(b.Patch)

	if aMajor == bMajor && aMinor == bMinor && aPatch == bPatch {
		return 0
	}

	if aMajor < bMajor {
		return -1
	}

	if aMinor < bMinor {
		return -1
	}

	if aPatch < bPatch {
		return -1
	}

	return 1
}

func MostRecentBranch(major, minor string, branches []*github.Branch) (Branch, error) {
	b := []Branch{}

	for _, v := range branches {
		version := strings.TrimSpace(strings.TrimPrefix(v.GetName(), "release-"))
		branchMajor, branchMinor, branchPatch := MajorMinorPatch(version)
		if major != branchMajor || minor != branchMinor {
			continue
		}
		if strings.Contains(v.GetName(), "+security") {
			continue
		}
		b = append(b, Branch{
			Name:  v.GetName(),
			SHA:   v.GetCommit().GetSHA(),
			Major: branchMajor,
			Minor: branchMinor,
			Patch: branchPatch,
		})
	}

	if len(b) == 0 {
		return Branch{}, errors.New("no release branch matches pattern")
	}

	slices.SortFunc(b, SortBranches)
	return b[len(b)-1], nil
}

func MajorMinorPatch(v string) (string, string, string) {
	matches := versions.SemverRegexp.FindStringSubmatch(strings.TrimPrefix(v, "v"))
	groups := make(map[string]string)
	for i, name := range versions.SemverRegexp.SubexpNames() {
		if i > 0 && i <= len(matches) {
			groups[name] = matches[i]
		}
	}

	return groups["major"], groups["minor"], groups["patch"]
}

func GetReleaseBranches(ctx context.Context, log *slog.Logger, client BranchClient, owner, repo string) ([]*github.Branch, error) {
	var (
		page     int
		count    = 100
		branches = []*github.Branch{}
	)

	for {
		log.Debug("listing branches", "page", page, "count", count)
		b, r, err := client.ListBranches(ctx, owner, repo, &github.BranchListOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: count,
			},
		})
		if err != nil {
			return nil, err
		}

		for _, branch := range b {
			// Filter here rather than sending the filtered request to the API,
			// as it is slower and it times out with 504 on repos with thousands of branches.
			if branch.Protected != nil && *branch.Protected {
				branches = append(branches, branch)
			}
		}

		if r.NextPage == 0 {
			break
		}
		page = r.NextPage
	}

	return branches, nil
}

func GetReleaseBranchByName(ctx context.Context, client BranchClient, owner, repo string, name string) (Branch, error) {
	b, _, err := client.GetBranch(ctx, owner, repo, name, true)
	return Branch{
		Name: b.GetName(),
		SHA:  b.GetCommit().GetSHA(),
	}, err
}
