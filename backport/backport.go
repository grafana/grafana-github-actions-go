package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v50/github"
)

var semverRegex = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>x|0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

type BackportOpts struct {
	PullRequestNumber int
	// SourceSHA is the commit hash that will be cherry-picked into a pull request targeting Target
	SourceSHA string

	// Target is the base branch of the backport pull request
	Target string

	// Labels are labels that will be added to the backport pull request
	Labels []*github.Label

	// IssueURL will set the "issue" field in the pull request
	IssueURL string

	TitleTemplate string

	Owner      string
	Repository string
}

func Run(ctx context.Context, command string, args ...string) (string, error) {
	return "", nil
}

func CherryPick(ctx context.Context, opts BackportOpts) error {
	branch := fmt.Sprintf("backport-%d-to-%s", opts.PullRequestNumber, opts.Target)
	if _, err := Run(ctx, "git", "fetch"); err != nil {
		return fmt.Errorf("error fetching: %w", err)
	}

	if _, err := Run(ctx, "git", "switch", "--create", branch, opts.Target); err != nil {
		return fmt.Errorf("error creating branch: %w", err)
	}

	stdout, err := Run(ctx, "git", "cherry-pick", "-x", opts.SourceSHA)
	if err != nil {
		if IsBettererConflict(ctx) {
			if _, err := Run(ctx, "git", "cherry-pick", "--continue"); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("error cherry-picking: %w", err)
	}

	return nil
}

func Backport(ctx context.Context, client *github.Client, opts BackportOpts) (string, error) {
	// 1. Run CLI commands to create a branch and cherry-pick
	//   * If the cherry-pick fails, write a comment in the source PR with instructions on manual backporting
	// 2. git push
	// 3. Open the pull request against the appropriate release branch

	if err := CherryPick(ctx, opts); err != nil {
	}

	return "", nil
}

func GetReleaseBranches(ctx context.Context, client *github.Client, owner, repo string) ([]string, error) {
	branches, _, err := client.Repositories.ListBranches(ctx, owner, repo, &github.BranchListOptions{
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

func BackportTargets(branches []string, labels []*github.Label) ([]string, error) {
	targets := []string{}
	for _, label := range labels {
		target, err := BackportTarget(label, branches)
		if err != nil {
			return nil, fmt.Errorf("error getting target for backport label '%s': %w", label, err)
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func MajorMinorPatch(v string) (string, string, string) {
	matches := semverRegex.FindStringSubmatch(strings.TrimPrefix(v, "v"))
	groups := make(map[string]string)
	for i, name := range semverRegex.SubexpNames() {
		if i > 0 && i <= len(matches) {
			groups[name] = matches[i]
		}
	}
	log.Println(groups)

	return groups["major"], groups["minor"], groups["patch"]
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

// BackportTarget finds the most appropriate base branch (target) given the backport label 'label'
// This function takes the label, like `backport v11.2.x`, and finds the most recent `release-` branch
// that matches the pattern.
func BackportTarget(label *github.Label, branches []string) (string, error) {
	version := strings.TrimPrefix(label.GetName(), "backport")
	labelString := strings.TrimSpace(version)
	major, minor, _ := MajorMinorPatch(labelString)

	return MostRecentBranch(major, minor, branches)
}
