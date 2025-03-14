package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/google/go-github/v50/github"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func main() {
	actions := []string{
		"github-release",
		"update-changelog",
		"auto-milestone",
		"community-release",
		"latest-release-branch",
		"backport",
		"bump-release",
		"migrate-open-prs",
	}

	var doTest bool
	var doBuild bool
	var doUpload bool
	var targetTag string

	pflag.BoolVar(&doTest, "do-test", false, "Execute tests")
	pflag.BoolVar(&doBuild, "do-build", false, "Execute builds")
	pflag.BoolVar(&doUpload, "do-upload", false, "Execute upload")
	pflag.StringVar(&targetTag, "target-tag", "dev", "Tag to upload an asset to")
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

	goContainer := client.Container().From(mustGetImage("golang")).
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOOS", "linux").
		WithEnvVariable("GOARCH", "amd64").
		WithMountedDirectory("/src", srcDir).
		WithMountedCache("/go/pkg/mod", goModCache).
		WithWorkdir("/src")

	if doTest {
		logger.Info().Msg("Running tests")
		if _, err := goContainer.WithExec([]string{"go", "test", "./...", "-v"}).Sync(ctx); err != nil {
			logger.Fatal().Err(err).Msg("Executing the tests failed")
		}
	}

	if doBuild {
		logger.Info().Msg("Building actions")
		for _, action := range actions {
			{
				logger := logger.With().Str("action", action).Logger()
				logger.Info().Msg("Building")

				goContainer = goContainer.WithWorkdir("/src/" + action).WithExec([]string{"go", "build"})
				if _, err := goContainer.Sync(ctx); err != nil {
					logger.Fatal().Msg("Building failed")
				}
			}
		}

		if doUpload {
			targetOwner := "grafana"
			targetRepo := "grafana-github-actions-go"
			logger.Info().Msg("Extracting binaries")
			tmpDir, err := os.MkdirTemp("", "go-actions-binaries")
			if err != nil {
				logger.Fatal().Err(err).Msg("Failed to create temporary folder for storing the binaries in")
			}
			defer os.RemoveAll(tmpDir)

			ghc := github.NewTokenClient(ctx, os.Getenv("GITHUB_TOKEN"))
			release, _, err := ghc.Repositories.GetReleaseByTag(ctx, targetOwner, targetRepo, targetTag)
			if err != nil {
				logger.Fatal().Err(err).Msgf("No release with the tag `%s` found", targetTag)
			}

			for _, action := range actions {
				{
					logger := logger.With().Str("action", action).Logger()
					if _, err := goContainer.File(fmt.Sprintf("/src/%s/%s", action, action)).Export(ctx, filepath.Join(tmpDir, action)); err != nil {
						logger.Fatal().Err(err).Msgf("Failed to export `%s` binary", action)
					}
					logger.Info().Msgf("Upload to %s release", targetTag)
					fp, err := os.Open(filepath.Join(tmpDir, action))
					if err != nil {
						logger.Fatal().Msg("Failed to open binary")
					}
					// Delete assets if they already exist
					assets, _, err := ghc.Repositories.ListReleaseAssets(ctx, targetOwner, targetRepo, release.GetID(), &github.ListOptions{})
					if err != nil {
						logger.Fatal().Err(err).Msg("Failed to get release assets")
					}
					for _, asset := range assets {
						if asset.GetName() == action {
							logger.Info().Msgf("Deleting old asset from release: %s", asset.GetName())
							if _, err := ghc.Repositories.DeleteReleaseAsset(ctx, targetOwner, targetRepo, asset.GetID()); err != nil {
								logger.Fatal().Err(err).Msg("Failed to delete release asset")
							}
						}
					}
					if _, _, err := ghc.Repositories.UploadReleaseAsset(ctx, targetOwner, targetRepo, release.GetID(), &github.UploadOptions{
						Name: action,
					}, fp); err != nil {
						logger.Fatal().Err(err).Msg("Failed to upload binary")
						fp.Close()
					}
					fp.Close()
				}
			}
			fmt.Println(release.GetAssetsURL())
		}
	}
}

func mustGetImage(basename string) string {
	fp, err := os.Open("Dockerfile.dagger")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FROM ") {
			elems := strings.SplitN(line, " ", 3)
			if len(elems) < 2 {
				continue
			}
			image := elems[1]
			imageParts := strings.SplitN(image, ":", 2)
			if len(imageParts) < 2 {
				continue
			}
			if imageParts[0] == basename {
				return image
			}
		}
	}
	panic(fmt.Errorf("no matching image found for %s", basename))
}
