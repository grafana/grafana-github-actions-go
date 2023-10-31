package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"

	"github.com/grafana/grafana-github-actions-go/pkg/changelog"
	"github.com/spf13/pflag"
)

func main() {
	var changelogFilePath string
	pflag.StringVar(&changelogFilePath, "changelog-file", "", "Path to the changelog file")
	pflag.Parse()

	if changelogFilePath == "" {
		panic(fmt.Errorf("no changelog file provided"))
	}

	ctx := context.Background()
	versions := []string{
		"10.2.0",
		"10.1.5", "10.1.4", "10.1.2", "10.1.1", "10.1.0",
		"10.0.9", "10.0.8", "10.0.6", "10.0.5", "10.0.4", "10.0.3", "10.0.4", "10.0.3", "10.0.2", "10.0.1", "10.0.0", "10.0.0-preview",
		"9.5.12", "9.5.10", "9.5.9", "9.5.8", "9.5.7", "9.5.6", "9.5.5", "9.5.3", "9.5.2", "9.5.1", "9.5.0",
		"9.4.17",
	}
	parse := changelog.NewParser()
	result := make(map[string][]changelog.Entry)
	for _, version := range versions {
		fp, err := os.Open(changelogFilePath)
		if err != nil {
			panic(err)
		}
		content, found, err := changelog.ExtractContentForVersion(ctx, fp, version, nil)
		if !found {
			fp.Close()
			panic(fmt.Errorf("no changelog found for %s", version))
		}
		fp.Close()
		if err != nil {
			panic(err)
		}
		sections, err := parse.Parse(ctx, bytes.NewBufferString(content))
		if err != nil {
			panic(err)
		}
		for _, section := range sections {
			res := result[section.Title]
			if res == nil {
				res = make([]changelog.Entry, 0, 10)
			}
			entries := make([]changelog.Entry, 0, len(section.Entries))
			for _, e := range section.Entries {
				e.Version = version
				entries = append(entries, e)
			}
			res = append(res, entries...)
			result[section.Title] = res
		}
	}

	out := csv.NewWriter(os.Stdout)
	out.Write([]string{"#title", "version", "issue", "section"})

	for section, entries := range result {
		seenTitles := make(map[string]struct{})
		filteredEntries := make([]changelog.Entry, 0, len(seenTitles))
		for _, e := range entries {
			if _, found := seenTitles[e.Title]; !found {
				filteredEntries = append(filteredEntries, e)
			}
			seenTitles[e.Title] = struct{}{}
		}
		sort.Slice(filteredEntries, func(i, j int) bool {
			return filteredEntries[i].Title < filteredEntries[j].Title
		})
		for _, entry := range filteredEntries {
			out.Write([]string{entry.Title, entry.Version, entry.Issue, section})
		}
	}
}
