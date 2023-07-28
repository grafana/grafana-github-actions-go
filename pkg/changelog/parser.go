package changelog

import (
	"bufio"
	"context"
	"io"
	"strings"
)

type Entry struct {
	Title string
}

type Section struct {
	Title   string
	Entries []Entry
}

// Parser provides functionality to parse the entries of a single
// version-changelog into a collection of sections. Note that the parsing
// cannot restore the complete state of the original ChangelogBody as the
// original serialization is lossful.
//
// Add this point only tickets in the "Bug fixes", "Features and enhancements",
// and "Plugin development fixes & changes" section work reliably.
type Parser struct{}

// NewParser generates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) rawParse(ctx context.Context, content io.Reader) ([]Section, error) {
	result := make([]Section, 0, 5)
	scanner := bufio.NewScanner(content)
	scanner.Split(bufio.ScanLines)
	inSection := false
	var currentSection *Section
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "### ") {
			inSection = true
			if currentSection != nil {
				result = append(result, *currentSection)
			}
			currentSection = &Section{
				Title:   strings.TrimPrefix(line, "### "),
				Entries: make([]Entry, 0, 10),
			}
			continue
		}
		if inSection && strings.HasPrefix(line, "- ") {
			// For the title we only care about anything that comes before the
			// link in the list item:
			elems := strings.SplitN(strings.TrimPrefix(line, "- "), "[", 2)
			title := elems[0]
			currentSection.Entries = append(currentSection.Entries, Entry{
				Title: strings.ReplaceAll(strings.TrimSpace(title), "*", ""),
			})
			continue
		}
	}
	if currentSection != nil {
		result = append(result, *currentSection)
	}
	return result, nil
}

func (p *Parser) Parse(ctx context.Context, content io.Reader) ([]Section, error) {
	return p.rawParse(ctx, content)
}
