package changelog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/grafana/grafana-github-actions-go/pkg/versions"
	"github.com/rs/zerolog"
)

type NoChangelogFound struct {
	Version string
}

func (e NoChangelogFound) Error() string {
	return fmt.Sprintf("no changelog found for `%s`", e.Version)
}

type LoaderOptions struct {
	RemoveHeading bool
}

// Loader can be used to retrieve an existing changelog from a GitHub
// repository structed similar to grafana/grafana.
type Loader struct {
	gh *github.Client
}

func NewLoader(gh *github.Client) *Loader {
	return &Loader{
		gh: gh,
	}
}

func (l *Loader) LoadContent(ctx context.Context, repoOwner string, repoName string, version string, opts *LoaderOptions) (string, error) {
	logger := zerolog.Ctx(ctx)
	vel := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(vel) < 1 {
		return "", fmt.Errorf("unsupported version provided")
	}
	majorVersion := vel[0]
	fileCandidates := []string{
		"CHANGELOG.md",
		fmt.Sprintf(".changelog-archive/CHANGELOG.%s.md", majorVersion),
	}

	versionBranch, err := versions.VersionBranch(version)
	if err != nil {
		return "", err
	}

	// Loads the CHANGELOG from the v branch instead of main
	for _, clPath := range fileCandidates {
		fc, _, resp, err := l.gh.Repositories.GetContents(ctx, repoOwner, repoName, clPath, &github.RepositoryContentGetOptions{
			Ref: versionBranch,
		})
		if resp.StatusCode == http.StatusNotFound {
			logger.Warn().Msgf("Changelog file not found: %s", clPath)
			continue
		}
		if err != nil {
			return "", err
		}
		rawContent, err := fc.GetContent()
		if err != nil {
			return "", err
		}
		buf := bytes.NewBufferString(rawContent)
		clContent, found, err := ExtractContentForVersion(ctx, buf, version, &ExtractContentOptions{
			RemoveHeadling: opts.RemoveHeading,
		})
		if err != nil {
			return "", err
		}
		if found {
			return clContent, nil
		}
	}

	return "", NoChangelogFound{Version: version}
}

type ExtractContentOptions struct {
	RemoveHeadling bool
}

// ExtractContentForVersion parses the provided fileContent for the content of
// a specific version and returns it if present.
func ExtractContentForVersion(ctx context.Context, fileContent io.Reader, version string, options *ExtractContentOptions) (string, bool, error) {
	opts := options
	if opts == nil {
		opts = &ExtractContentOptions{
			RemoveHeadling: false,
		}
	}
	v := strings.TrimPrefix(version, "v")
	output := strings.Builder{}
	sc := bufio.NewScanner(fileContent)
	sc.Split(bufio.ScanLines)
	inVersion := false
	versionStart := fmt.Sprintf("<!-- %s START -->", v)
	versionEnd := fmt.Sprintf("<!-- %s END -->", v)
	nonEmptyLines := 0
	for sc.Scan() {
		line := sc.Text()
		if !inVersion && line == versionStart {
			inVersion = true
			continue
		}
		if inVersion {
			if line == versionEnd {
				result := strings.TrimSpace(output.String())
				return result, true, nil
			}
			// Remove the first non-empty line that starts with a # if
			// requested:
			if opts.RemoveHeadling && strings.HasPrefix(line, "# ") && nonEmptyLines == 0 {
				continue
			}
			output.WriteString(line)
			output.WriteRune('\n')
			if line != "" {
				nonEmptyLines += 1
			}
		}

	}
	return output.String(), false, nil
}
