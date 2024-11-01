package main

import (
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
)

func TestMostRecentBranch(t *testing.T) {
	assertError := func(t *testing.T, major, minor string, branches []string) {
		t.Helper()
		b, err := MostRecentBranch(major, minor, branches)
		assert.Error(t, err)
		assert.Empty(t, b)
	}

	assertBranch := func(t *testing.T, major, minor string, branches []string, branch string) {
		t.Helper()
		b, err := MostRecentBranch(major, minor, branches)
		assert.NoError(t, err)
		assert.Equal(t, branch, b)
	}
	branches := []string{
		"release-11.0.1",
		"release-1.2.3",
		"release-11.0.1+security-01",
		"release-10.0.0",
		"release-10.2.3",
		"release-10.2.4",
		"release-10.2.4+security-01",
		"release-12.0.3",
		"release-12.1.3",
		"release-12.0.15",
		"release-12.1.15",
		"release-12.2.12",
	}

	assertError(t, "3", "2", branches)
	assertError(t, "4", "0", branches)
	assertError(t, "13", "0", branches)
	assertError(t, "10", "5", branches)
	assertError(t, "11", "8", branches)
	assertBranch(t, "11", "0", branches, "release-11.0.1")
	assertBranch(t, "12", "1", branches, "release-12.1.15")
	assertBranch(t, "12", "0", branches, "release-12.0.15")
	assertBranch(t, "1", "2", branches, "release-1.2.3")
	assertBranch(t, "10", "2", branches, "release-10.2.4")
}

func TestBackportTarget(t *testing.T) {
	assertError := func(t *testing.T, label *github.Label, branches []string) {
		t.Helper()
		b, err := BackportTarget(label, branches)
		assert.Error(t, err)
		assert.Empty(t, b)
	}

	assertBranch := func(t *testing.T, label *github.Label, branches []string, branch string) {
		t.Helper()
		b, err := BackportTarget(label, branches)
		assert.NoError(t, err)
		assert.Equal(t, branch, b)
	}

	branches := []string{
		"release-11.0.1",
		"release-1.2.3",
		"release-11.0.1+security-01",
		"release-10.0.0",
		"release-10.2.3",
		"release-10.2.4",
		"release-10.2.4+security-01",
		"release-12.0.3",
		"release-12.1.3",
		"release-12.0.15",
		"release-12.1.15",
		"release-12.2.12",
	}

	assertError(t, &github.Label{
		Name: github.String("backport v3.2.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v4.0.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v13.0.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v10.5.x"),
	}, branches)
	assertError(t, &github.Label{
		Name: github.String("backport v11.8.x"),
	}, branches)
	assertBranch(t, &github.Label{
		Name: github.String("backport v11.0.x"),
	}, branches, "release-11.0.1")
	assertBranch(t, &github.Label{
		Name: github.String("backport v12.1.x"),
	}, branches, "release-12.1.15")
	assertBranch(t, &github.Label{
		Name: github.String("backport v12.0.x"),
	}, branches, "release-12.0.15")
	assertBranch(t, &github.Label{
		Name: github.String("backport v1.2.x"),
	}, branches, "release-1.2.3")
	assertBranch(t, &github.Label{
		Name: github.String("backport v10.2.x"),
	}, branches, "release-10.2.4")
}
