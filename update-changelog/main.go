package main

import (
	"context"
	"fmt"
	"grafana-github-actions-go/internal/changelog"
	"grafana-github-actions-go/internal/toolkit"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func main() {
	var version string
	pflag.StringVar(&version, "version", "", "Version to target")
	pflag.Parse()

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx := logger.WithContext(context.Background())
	tk, err := toolkit.Init(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize toolkit")
	}

	if version == "" {
		version = tk.GetInput("version", nil)
	}
	if version == "" {
		logger.Fatal().Msg("No version specified")
	}

	body, err := changelog.Build(ctx, version, tk)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to build changelog")
	}
	fmt.Println(body.ToMarkdown())
}
