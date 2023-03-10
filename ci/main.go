package main

import (
	"context"
	"os"

	"dagger.io/dagger"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func main() {
	actions := []string{"update-changelog"}

	var doTest bool
	var doBuild bool

	pflag.BoolVar(&doTest, "do-test", false, "Execute tests")
	pflag.BoolVar(&doBuild, "do-build", false, "Execute builds")
	pflag.Parse()

	ctx := context.Background()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to dagger")
	}
	defer client.Close()

	srcDir := client.Host().Directory(".")

	goModCache := client.CacheVolume("gomodcache")

	goContainer := client.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
	}).From("golang:1.20.2").
		WithEnvVariable("CGO_ENABLED", "0").
		WithMountedDirectory("/src", srcDir).
		WithMountedCache("/go/pkg/mod", goModCache).
		WithWorkdir("/src")

	if doTest {
		logger.Info().Msg("Running tests")
		if _, err := goContainer.WithExec([]string{"go", "test", "./...", "-v"}).ExitCode(ctx); err != nil {
			logger.Fatal().Err(err).Msg("Executing the tests failed")
		}
	}

	if doBuild {
		logger.Info().Msg("Building actions")
		for _, action := range actions {
			{
				logger := logger.With().Str("action", action).Logger()
				logger.Info().Msg("Building")
				if _, err := goContainer.WithWorkdir("/src/" + action).WithExec([]string{"go", "build"}).ExitCode(ctx); err != nil {
					logger.Fatal().Msg("Building failed")
				}
			}
		}
	}
}
