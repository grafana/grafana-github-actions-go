package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/google/go-github/v50/github"
)

type CommentData struct {
	Target                  string
	Error                   string
	BackportBranch          string
	SourceSHA               string
	SourcePullRequestNumber int
	Body                    string
}

type FailureOpts struct {
	BackportOpts
	Error error
}

func CommentFailure(ctx context.Context, client *github.Client, opts FailureOpts) error {
	var (
		branch   = BackportBranch(opts.PullRequestNumber, opts.Target)
		bodyText = opts.SourceBody
	)

	if bodyText == "" {
		bodyText = fmt.Sprintf("backport %d to %s", opts.PullRequestNumber, branch)
	}

	data := CommentData{
		Target:                  opts.Target,
		Error:                   opts.Error.Error(),
		BackportBranch:          branch,
		SourceSHA:               opts.SourceSHA,
		SourcePullRequestNumber: opts.PullRequestNumber,
		Body:                    bodyText,
	}

	tmpl, err := template.New("").Parse(commentTemplate)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(nil)
	if err := tmpl.Execute(body, data); err != nil {
		return err
	}
	_, _, err = client.PullRequests.CreateComment(ctx, opts.Owner, opts.Repository, opts.PullRequestNumber, &github.PullRequestComment{
		Body: github.String(body.String()),
	})
	if err != nil {
		return err
	}

	return nil
}
