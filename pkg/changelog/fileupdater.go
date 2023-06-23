package changelog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/coreos/go-semver/semver"
)

var versionEndLinePattern = regexp.MustCompile("<!-- (.*) END -->")
var versionStartLinePattern = regexp.MustCompile("<!-- (.*) START -->")

func UpdateFileAtPath(ctx context.Context, file string, rendered string, body *ChangelogBody) error {
	input, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()
	output := bytes.Buffer{}
	if err := UpdateFile(ctx, &output, input, rendered, body); err != nil {
		return err
	}
	return os.WriteFile(file, output.Bytes(), 0o644)
}

// UpdateFile receives the original changelog data via the `in` parameter and
// writes it back to the `out` parameter with the `body` behing inserted.
func UpdateFile(ctx context.Context, out io.Writer, in io.Reader, rendered string, body *ChangelogBody) error {
	newVersion := semver.New(body.Version)
	scanner := bufio.NewScanner(in)
	scanner.Split(bufio.ScanLines)
	var inserted bool

	insertEntry := func() {
		if inserted {
			return
		}
		out.Write([]byte("<!-- "))
		out.Write([]byte(body.Version))
		out.Write([]byte(" START -->\n\n"))

		out.Write([]byte(rendered))

		out.Write([]byte("<!-- "))
		out.Write([]byte(body.Version))
		out.Write([]byte(" END -->\n"))
		inserted = true
	}

	lines := make([]string, 0, 50)
	insertAfterIdx := -1
	replaceAfterIdx := -1
	replaceBeforeIdx := -1
	idx := 0
	for scanner.Scan() {
		line := string(scanner.Bytes())
		lines = append(lines, line)
		if match := versionStartLinePattern.FindStringSubmatch(line); len(match) > 1 {
			version := match[1]
			v := semver.New(version)
			if v.Equal(*newVersion) {
				replaceAfterIdx = idx
			}
		}
		if match := versionEndLinePattern.FindStringSubmatch(line); len(match) > 1 {
			version := match[1]
			v := semver.New(version)
			if v.Equal(*newVersion) {
				replaceBeforeIdx = idx
			} else if !v.LessThan(*newVersion) {
				insertAfterIdx = idx
			}
		}
		idx++
	}

	for idx, line := range lines {
		// If we are in the range that we want to replace, skip those lines:
		if replaceAfterIdx != -1 && replaceBeforeIdx != -1 && idx <= replaceBeforeIdx && idx >= replaceAfterIdx {
			continue
		}
		if replaceAfterIdx != -1 && replaceAfterIdx == idx {
			insertEntry()
		}
		if insertAfterIdx < idx {
			insertEntry()
		}

		out.Write([]byte(line))
		out.Write([]byte("\n"))
		if insertAfterIdx == idx {
			insertEntry()
		}
	}
	if !inserted {
		insertEntry()
	}
	return nil
}
