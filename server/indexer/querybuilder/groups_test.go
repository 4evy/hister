package querybuilder

import (
	"errors"
	"slices"
	"testing"

	"github.com/blevesearch/bleve/v2"
)

func TestBooleanGroups(t *testing.T) {
	idx := newRankingTestIndex(t, map[string]map[string]string{
		"a": {"title": "privacy policy", "text": "guide", "domain": "a.test", "url": "https://a.test/guide"},
		"b": {"title": "privacy complete policy", "text": "reference", "domain": "b.test", "url": "https://b.test/reference"},
		"c": {"title": "data protection", "text": "guide", "domain": "c.test", "url": "https://c.test/guide"},
		"d": {"title": "security", "text": "privacy policy", "domain": "d.test"},
		"e": {"title": "privacy", "text": "guide", "domain": "e.test"},
	})
	for _, test := range []struct {
		input string
		want  []string
	}{
		{`("privacy policy"|"data protection")`, []string{"a", "c", "d"}},
		{`title:("privacy policy" | "data protection")`, []string{"a", "c"}},
		{`title:(privacy| security)`, []string{"a", "b", "d", "e"}},
		{`title:("privacy policy" OR security)`, []string{"a", "d"}},
		{`(domain:a.test|-domain:b.test)`, []string{"a", "c", "d", "e"}},
		{`-(domain:a.test|domain:b.test)`, []string{"c", "d", "e"}},
		{`domain:-(a.test|b.test)`, []string{"c", "d", "e"}},
		{`-domain:(a.test|b.test)`, []string{"c", "d", "e"}},
		{`(domain:a.test | (domain:b.test | domain:c.test))`, []string{"a", "b", "c"}},
		{`domain:a.test OR domain:b.test AND text:guide`, []string{"a"}},
		{`domain:a.test|domain:b.test text:guide`, []string{"a"}},
		{`(domain:a.test OR domain:b.test) AND text:guide`, []string{"a"}},
		{`NOT domain:b.test AND text:guide`, []string{"a", "c", "e"}},
		{`(title:privacy text:guide) OR title:security`, []string{"a", "d", "e"}},
		{`(-domain:a.test -domain:b.test)`, []string{"c", "d", "e"}},
		{`title:(security OR text:reference)`, []string{"b", "d"}},
		{`title:(privacy AND (policy OR protection))`, []string{"a", "b"}},
		{`(privacy guide)`, []string{"a", "e"}},
		{`(domain:b.test OR *) text:guide`, []string{"a", "c", "e"}},
		{`(* OR title:privacy) -*`, nil},
		{`(domain:a.test OR -*)`, []string{"a"}},
		{`NOT NOT domain:a.test`, []string{"a"}},
		{`(url_re:"https://(a|b)\.test/" OR domain:c.test)`, []string{"a", "b", "c"}},
		{`(url_re:https://(a|b)\.test/ OR domain:c.test)`, []string{"a", "b", "c"}},
		{`title:(privacy AND NOT policy)`, []string{"e"}},
		{`title:(privacy -policy)`, []string{"e"}},
	} {
		t.Run(test.input, func(t *testing.T) {
			assertGroupMatches(t, idx, test.input, test.want)
		})
	}
}

func TestGroupQuotedAndEscapedValues(t *testing.T) {
	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping.AddFieldMappingsAt("label", bleve.NewKeywordFieldMapping())
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = idx.Close() })
	for id, label := range map[string]string{
		"pipe": "a|b", "parens": "a(b)", "negative": "-archived", "operator": "AND",
		"quote": `say "hello"`, "backslash": `a\b`, "field": "domain:a.test",
	} {
		if err := idx.Index(id, map[string]string{"label": label}); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		input string
		want  []string
	}{
		{`label:("a|b"|"a(b)")`, []string{"parens", "pipe"}},
		{`label:(a\|b|a\(b\))`, []string{"parens", "pipe"}},
		{`label:("-archived" OR "AND")`, []string{"negative", "operator"}},
		{`label:(\-archived OR \AND)`, []string{"negative", "operator"}},
		{`label:("say \"hello\"" OR "a\\b")`, []string{"backslash", "quote"}},
		{`label:("domain:a.test")`, []string{"field"}},
	} {
		t.Run(test.input, func(t *testing.T) {
			assertGroupMatches(t, idx, test.input, test.want)
		})
	}
}

func TestNestedGroupsValidateURLRegexps(t *testing.T) {
	for _, input := range []string{
		`domain:a.test OR (title:privacy AND (url_re:"[" OR title:security))`,
		`NOT (title:privacy OR url_re:"(")`,
	} {
		if _, err := BuildValidated(input); !errors.Is(err, ErrInvalidRegexp) {
			t.Errorf("BuildValidated(%q) error = %v, want ErrInvalidRegexp", input, err)
		}
	}
}

func TestGroupedURLRegexpPreservesLiteralSyntax(t *testing.T) {
	idx := newRankingTestIndex(t, map[string]map[string]string{
		"dash":   {"url": "https://example.com/a-b"},
		"parens": {"url": "https://example.com/a(b)"},
		"slash":  {"url": `https://example.com/a\`},
		"other":  {"url": "https://example.org/", "domain": "example.org"},
	})
	for _, test := range []struct {
		input string
		want  []string
	}{
		{`url_re:-b`, []string{"dash"}},
		{`(url_re:-b OR domain:example.org)`, []string{"dash", "other"}},
		{`(url_re:"a\\" OR domain:example.org)`, []string{"other", "slash"}},
		{`(url_re:[()] OR domain:example.org)`, []string{"other", "parens"}},
		{`url:https://example.com/a(b)`, []string{"parens"}},
		{`(url:https://example.com/a(b) OR domain:example.org)`, []string{"other", "parens"}},
	} {
		t.Run(test.input, func(t *testing.T) {
			assertGroupMatches(t, idx, test.input, test.want)
		})
	}
}

func assertGroupMatches(t *testing.T, idx bleve.Index, input string, want []string) {
	t.Helper()
	q, err := BuildValidated(input)
	if err != nil {
		t.Fatal(err)
	}
	result, err := idx.Search(bleve.NewSearchRequest(q))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, hit := range result.Hits {
		got = append(got, hit.ID)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) || result.Total != uint64(len(want)) {
		t.Fatalf("Search(%q) = %v (total %d), want %v", input, got, result.Total, want)
	}
}
