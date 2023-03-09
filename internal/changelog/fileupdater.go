package changelog

import (
	"bufio"
	"context"
	"io"
	"regexp"

	"github.com/coreos/go-semver/semver"
)

var versionEndLinePattern = regexp.MustCompile("<!-- (.*) END -->")

// UpdateFile receives the original changelog data via the `in` parameter and
// writes it back to the `out` parameter with the `body` behing inserted.
func UpdateFile(ctx context.Context, out io.Writer, in io.Reader, body *ChangelogBody) error {
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

		out.Write([]byte(body.ToMarkdown()))

		out.Write([]byte("<!-- "))
		out.Write([]byte(body.Version))
		out.Write([]byte(" END -->\n"))
		inserted = true
	}

	lines := make([]string, 0, 50)
	insertAfterIdx := -1
	idx := 0
	for scanner.Scan() {
		line := string(scanner.Bytes())
		lines = append(lines, line)
		match := versionEndLinePattern.FindStringSubmatch(line)
		if len(match) > 1 {
			version := match[1]
			v := semver.New(version)
			if v.LessThan(*newVersion) {
			} else {
				insertAfterIdx = idx
			}
		}
		idx++
	}

	for idx, line := range lines {
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
