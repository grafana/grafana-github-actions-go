package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

type PullRequestInfo struct {}
