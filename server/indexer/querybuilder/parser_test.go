package querybuilder

import (
	"errors"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	type wantToken struct {
		kind TokenType
		text string
	}
	for _, test := range []struct {
		input string
		want  []wantToken
	}{
		{"", nil},
		{" \t ", nil},
		{"سلام", []wantToken{{TokenWord, "سلام"}}},
		{"hello", []wantToken{{TokenWord, "hello"}}},
		{"foo bar baz", []wantToken{{TokenWord, "foo"}, {TokenWord, "bar"}, {TokenWord, "baz"}}},
		{`"hello world"`, []wantToken{{TokenQuoted, "hello world"}}},
		{`"say \"hello\""`, []wantToken{{TokenQuoted, `say "hello"`}}},
		{"-hello", []wantToken{{TokenNot, "-"}, {TokenWord, "hello"}}},
		{"-", []wantToken{{TokenWord, "-"}}},
		{"title:golang", []wantToken{{TokenField, "title"}, {TokenWord, "golang"}}},
		{`title:"hello world"`, []wantToken{{TokenField, "title"}, {TokenQuoted, "hello world"}}},
		{"(hello|world)", []wantToken{{TokenLeftParen, "("}, {TokenWord, "hello"}, {TokenOr, "|"}, {TokenWord, "world"}, {TokenRightParen, ")"}}},
		{"( hello | مرحبا )", []wantToken{{TokenLeftParen, "("}, {TokenWord, "hello"}, {TokenOr, "|"}, {TokenWord, "مرحبا"}, {TokenRightParen, ")"}}},
		{"(only)", []wantToken{{TokenLeftParen, "("}, {TokenWord, "only"}, {TokenRightParen, ")"}}},
		{"AND OR NOT", []wantToken{{TokenAnd, "AND"}, {TokenOr, "OR"}, {TokenNot, "NOT"}}},
		{"and or not", []wantToken{{TokenWord, "and"}, {TokenWord, "or"}, {TokenWord, "not"}}},
		{`\AND a\|b`, []wantToken{{TokenWord, "AND"}, {TokenWord, "a|b"}}},
		{`"domain:example.com"`, []wantToken{{TokenQuoted, "domain:example.com"}}},
		{`url_re:^https?://example\.com/(private|login)$`, []wantToken{{TokenField, "url_re"}, {TokenWord, `^https?://example\.com/(private|login)$`}}},
		{`url_re:"^file:///.*/My Documents/.*\.pdf$"`, []wantToken{{TokenField, "url_re"}, {TokenQuoted, `^file:///.*/My Documents/.*\.pdf$`}}},
	} {
		t.Run(test.input, func(t *testing.T) {
			tokens, err := Tokenize(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(tokens) != len(test.want) {
				t.Fatalf("Tokenize(%q) = %v, want %v", test.input, tokens, test.want)
			}
			for i, want := range test.want {
				if tokens[i].Type != want.kind || tokens[i].Value != want.text {
					t.Errorf("token %d = %v, want %v", i, tokens[i], want)
				}
				if tokens[i].start >= tokens[i].end || tokens[i].end > len([]rune(test.input)) {
					t.Errorf("invalid source span for token %d: %d..%d", i, tokens[i].start, tokens[i].end)
				}
			}
		})
	}
}

func TestExpressionRejectsIncompleteGroupsAndOperators(t *testing.T) {
	for _, input := range []string{"(foo|bar", "foo)", "()", "(foo|)", "(|bar)", "foo OR", "AND foo", "NOT", "foo AND OR bar"} {
		tokens, err := Tokenize(input)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseExpression(tokens); err == nil {
			t.Errorf("parseExpression(%q) must reject incomplete syntax", input)
		}
	}
}

func TestExpressionBoundsNesting(t *testing.T) {
	for _, input := range []string{
		strings.Repeat("(", maxExpressionDepth+1) + "needle" + strings.Repeat(")", maxExpressionDepth+1),
		strings.Repeat("NOT ", maxExpressionDepth+1) + "needle",
	} {
		if _, err := BuildValidated(input); !errors.Is(err, errExpressionDepth) {
			t.Errorf("deep query error = %v, want nesting limit", err)
		}
	}
}
