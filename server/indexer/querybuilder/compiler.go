package querybuilder

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

func buildExpression(root *expression) query.Query {
	root = simplifyExpression(root, "")
	compiled := compileExpression(root, "", time.Now())
	switch compiled.(type) {
	case *query.MatchAllQuery, *query.MatchNoneQuery:
		return compiled
	}
	q, ok := compiled.(*query.BooleanQuery)
	if !ok {
		q = query.NewBooleanQuery([]query.Query{compiled}, nil, nil)
	}
	// Bonuses only apply to positive free text. Boolean operators and field
	// syntax are matching instructions, not relevance text.
	if text, ok := rankingText(root); ok {
		switch root.kind {
		case expressionAnd:
			q.AddShould(createCombinedMatchQuery(text, 2))
		case expressionTerm:
			uq := bleve.NewRegexpQuery(fmt.Sprintf("https?://(www\\.)?%s[^/]*/", regexp.QuoteMeta(strings.ToLower(text))))
			uq.SetField("url")
			uq.SetBoost(100)
			q.AddShould(uq)
		}
	}
	return q
}

func compileExpression(e *expression, field string, now time.Time) query.Query {
	switch e.kind {
	case expressionAll:
		return query.NewMatchAllQuery()
	case expressionNone:
		return query.NewMatchNoneQuery()
	case expressionTerm, expressionPhrase:
		return buildValueQuery(field, e.value, e.kind == expressionPhrase, now)
	case expressionField:
		return compileExpression(e.children[0], e.value, now)
	case expressionNot:
		return query.NewBooleanQuery(nil, nil, []query.Query{compileExpression(e.children[0], field, now)})
	case expressionOr:
		children := make([]query.Query, len(e.children))
		for i, child := range e.children {
			children[i] = compileExpression(child, field, now)
		}
		return bleve.NewDisjunctionQuery(children...)
	case expressionAnd:
		var required, excluded []query.Query
		for _, child := range e.children {
			if child.kind == expressionNot {
				excluded = append(excluded, compileExpression(child.children[0], field, now))
			} else {
				required = append(required, compileExpression(child, field, now))
			}
		}
		return query.NewBooleanQuery(required, nil, excluded)
	default:
		return query.NewMatchNoneQuery()
	}
}

func rankingText(e *expression) (string, bool) {
	switch e.kind {
	case expressionTerm, expressionPhrase:
		return e.value, true
	case expressionAnd:
		parts := make([]string, len(e.children))
		for i, child := range e.children {
			var ok bool
			parts[i], ok = rankingText(child)
			if !ok {
				return "", false
			}
		}
		return strings.Join(parts, " "), true
	default:
		return "", false
	}
}
