package querybuilder

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/blevesearch/bleve/v2/search/query"
)

func TestInvalidQueryDiagnostics(t *testing.T) {
	for _, tt := range []struct {
		input    string
		message  string
		position int
	}{
		{`"unfinished`, "unclosed quote", 1},
		{`世界 title:"unfinished`, "unclosed quote", 10},
		{`"escaped\"`, "unclosed quote", 1},
		{"(foo OR bar", "unclosed group", 1},
		{"foo)", "unexpected closing parenthesis", 4},
		{"()", "empty group", 1},
		{"foo OR", "expected a term after OR", 5},
		{"foo AND   ", "expected a term after AND", 5},
		{"NOT", "expected a term after NOT", 1},
		{"(foo|)", "expected a term after |", 5},
		{"AND foo", "expected a term", 1},
		{"foo AND OR bar", "expected a term after AND", 5},
		{"title:", "expected a value after title:", 1},
		{"(title:)", "expected a value after title:", 2},
		{"title:()", "empty group", 7},
		{`title:""`, "empty quoted value", 7},
		{"url_re:[", "invalid URL regexp", 8},
	} {
		t.Run(tt.input, func(t *testing.T) {
			q, err := BuildValidated(tt.input)
			if err == nil {
				t.Fatalf("BuildValidated(%q) = %T without an error", tt.input, q)
			}
			if q != nil || !strings.Contains(err.Error(), tt.message) || !strings.Contains(err.Error(), fmt.Sprintf("at character %d:", tt.position)) {
				t.Errorf("BuildValidated = %T, %v; want %q at character %d", q, err, tt.message, tt.position)
			}
			var diagnostic *QueryError
			if !errors.Is(err, ErrInvalidQuery) || !errors.As(err, &diagnostic) || diagnostic.Position != tt.position {
				t.Errorf("invalid query lacks a typed diagnostic: %v", err)
			}
			if _, ok := Build(tt.input).(*query.MatchNoneQuery); !ok {
				t.Error("Build must match nothing when it cannot report an invalid query")
			}
		})
	}
}

func TestQueryDiagnosticsPreserveLiteralSyntax(t *testing.T) {
	for _, input := range []string{
		`"AND"`, `\OR`, `\(foo\)`, `hello\|world`, `"say \"hello\""`,
		`https://example.com`, `unknown:value`, `title:OR`, `-`, `path\`,
		`url:https://example.com/wiki/Go_(language)`, `url_re:(foo|bar)`,
		`title:(foo OR bar) NOT archived`, "", "  ",
	} {
		if _, err := BuildValidated(input); err != nil {
			t.Errorf("BuildValidated(%q): %v", input, err)
		}
	}
}
