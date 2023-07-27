package changelog

import (
	"context"
	"io"
	"io/ioutil"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
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

func (p *Parser) Parse(ctx context.Context, content io.Reader) ([]Section, error) {
	result := make([]Section, 0, 5)
	mdParser := goldmark.DefaultParser()
	rawContent, err := ioutil.ReadAll(content)
	if err != nil {
		return nil, err
	}
	node := mdParser.Parse(text.NewReader(rawContent))
	inHeading := false
	var currentSection *Section
	err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch e := n.(type) {
		case *ast.Heading:
			if e.Level == 3 {
				inHeading = entering
			}
		case *ast.Text:
			if inHeading && !entering {
				title := string(e.Text(rawContent))
				if currentSection != nil {
					result = append(result, *currentSection)
				}
				currentSection = &Section{
					Title:   title,
					Entries: make([]Entry, 0, 10),
				}
			}
		case *ast.ListItem:
			if currentSection != nil && !entering {
				title := p.getTitle(rawContent, e)
				currentSection.Entries = append(currentSection.Entries, Entry{Title: title})
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}
	if currentSection != nil {
		result = append(result, *currentSection)
	}
	return result, nil
}

func (p *Parser) getTitle(source []byte, listItem ast.Node) string {
	textBlock := listItem.FirstChild()
	elements := make([]string, 0, 2)
	ast.Walk(textBlock, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if c.Parent() == textBlock && !entering {
			if c.Kind() == ast.KindLink {
				return ast.WalkStop, nil
			}
			text := c.Text(source)
			elements = append(elements, string(text))
		}
		return ast.WalkContinue, nil
	})
	title := strings.TrimSpace(strings.Join(elements, ""))
	return title
}
