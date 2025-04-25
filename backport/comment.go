package main

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/google/go-github/v50/github"
)

type CommentData struct {
	BackportTitle           string
	Target                  string
	Error                   string
	BackportBranch          string
	SourceSHA               string
	SourcePullRequestNumber int
	Body                    string
	Labels                  []string
}

type FailureOpts struct {
	BackportOpts
	Error error
}

func CommentFailure(ctx context.Context, client CommentClient, opts FailureOpts) error {
	var (
		branch   = BackportBranch(opts.PullRequestNumber, opts.Target.Name)
		bodyText = opts.SourceBody
	)

	if bodyText == "" {
		bodyText = fmt.Sprintf("backport %d to %s", opts.PullRequestNumber, branch)
	}

	labels := make([]string, len(opts.Labels))
	for i, v := range opts.Labels {
		labels[i] = v.GetName()
	}

	data := CommentData{
		BackportTitle:           fmt.Sprintf("[%s] %s", opts.Target.Name, opts.SourceTitle),
		Target:                  opts.Target.Name,
		Error:                   opts.Error.Error(),
		BackportBranch:          branch,
		SourceSHA:               opts.SourceSHA,
		SourcePullRequestNumber: opts.PullRequestNumber,
		Body:                    bodyText,
		Labels:                  labels,
	}

	tmpl, err := template.New("").Parse(commentTemplate)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(nil)
	if err := tmpl.Execute(body, data); err != nil {
		return err
	}
	_, _, err = client.CreateComment(ctx, opts.Owner, opts.Repository, opts.PullRequestNumber, &github.IssueComment{
		Body: github.String(body.String()),
	})
	if err != nil {
		return fmt.Errorf("error creating comment for error '%s': %w", opts.Error.Error(), err)
	}

	return nil
}
