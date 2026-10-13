package indexer

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/blevesearch/bleve/v2/search/highlight"
	htmlFormatter "github.com/blevesearch/bleve/v2/search/highlight/format/html"
	simpleFragmenter "github.com/blevesearch/bleve/v2/search/highlight/fragmenter/simple"
	simpleHighlighter "github.com/blevesearch/bleve/v2/search/highlight/highlighter/simple"
)

const snippetHighlighterType = "hister_snippet"

func init() {
	// Register constructors before searches can read Bleve's global registry.
	if err := registry.RegisterHighlighter(snippetHighlighterType, newSnippetHighlighter); err != nil {
		panic(err)
	}
}

func snippetStyles(length int) (map[string]string, error) {
	if length < 1 {
		return nil, fmt.Errorf("app.snippet_length must be a positive integer")
	}
	styles := make(map[string]string, 3)
	for _, style := range []string{"HTML", "text", "tui"} {
		// Cached highlighters are immutable and keyed by size so independent
		// indexers never change each other's snippets.
		name := fmt.Sprintf("%s_%s_%d", snippetHighlighterType, style, length)
		_, err := bleve.Config.Cache.DefineHighlighter(name, map[string]any{
			"type": snippetHighlighterType, "style": style, "size": length,
		})
		if err != nil {
			// Another indexer may already have defined the same highlighter.
			if _, lookupErr := bleve.Config.Cache.HighlighterNamed(name); lookupErr != nil {
				return nil, err
			}
		}
		styles[style] = name
	}
	return styles, nil
}

func newSnippetHighlighter(options map[string]any, _ *registry.Cache) (highlight.Highlighter, error) {
	length, ok := options["size"].(int)
	if !ok || length < 1 {
		return nil, fmt.Errorf("snippet highlighter requires a positive size")
	}
	var formatter highlight.FragmentFormatter
	switch options["style"] {
	case "HTML":
		formatter = htmlFormatter.NewFragmentFormatter("<mark>", "</mark>")
	case "text":
		formatter = &lipglossFormatter{style: lipgloss.NewStyle().Reverse(true)}
	case "tui":
		formatter = &lipglossFormatter{style: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)}
	default:
		return nil, fmt.Errorf("unknown snippet highlight style %v", options["style"])
	}
	return simpleHighlighter.NewHighlighter(simpleFragmenter.NewFragmenter(length), formatter, simpleHighlighter.DefaultSeparator), nil
}

type lipglossFormatter struct {
	style lipgloss.Style
}

func (f *lipglossFormatter) Format(fragment *highlight.Fragment, orderedTermLocations highlight.TermLocations) string {
	var sb strings.Builder
	curr := fragment.Start
	for _, tl := range orderedTermLocations {
		if tl == nil || !tl.ArrayPositions.Equals(fragment.ArrayPositions) || tl.Start < curr || tl.End > fragment.End {
			continue
		}
		sb.WriteString(string(fragment.Orig[curr:tl.Start]))
		sb.WriteString(f.style.Render(string(fragment.Orig[tl.Start:tl.End])))
		curr = tl.End
	}
	sb.WriteString(string(fragment.Orig[curr:fragment.End]))
	return sb.String()
}

// truncateText trims s to at most maxRunes runes, appending "…" when cut.
func truncateText(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}
