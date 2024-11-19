package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sethvargo/go-githubactions"
)

func UnmarshalEventData(ctx *githubactions.GitHubContext, v any) error {
	if ctx.EventPath == "" {
		return fmt.Errorf("event path is empty")
	}

	eventData, err := os.ReadFile(ctx.EventPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not read event file: %w", err)
	}

	if eventData == nil {
		return nil
	}

	return json.Unmarshal(eventData, v)
}
