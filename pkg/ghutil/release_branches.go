package ghutil

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/versions"
)

type BranchClient interface {
	ListBranches(ctx context.Context, owner string, repo string, opts *github.BranchListOptions) ([]*github.Branch, *github.Response, error)
}

type Branch struct {
	Name  string
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

func MostRecentBranch(major, minor string, branches []string) (string, error) {
	b := []Branch{}

	for _, v := range branches {
		version := strings.TrimSpace(strings.TrimPrefix(v, "release-"))
		branchMajor, branchMinor, branchPatch := MajorMinorPatch(version)
		if major != branchMajor || minor != branchMinor {
			continue
		}
		if strings.Contains(v, "+security") {
			continue
		}
		b = append(b, Branch{
			Name:  v,
			Major: branchMajor,
			Minor: branchMinor,
			Patch: branchPatch,
		})
	}

	if len(b) == 0 {
		return "", errors.New("no release branch matches pattern")
	}

	slices.SortFunc(b, SortBranches)
	return b[len(b)-1].Name, nil
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

func GetReleaseBranches(ctx context.Context, client BranchClient, owner, repo string) ([]string, error) {
	branches, _, err := client.ListBranches(ctx, owner, repo, &github.BranchListOptions{
		Protected: github.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	str := make([]string, len(branches))
	for i, v := range branches {
		str[i] = v.GetName()
	}

	return str, nil
}
