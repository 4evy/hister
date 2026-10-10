// SPDX-License-Identifier: AGPL-3.0-or-later

package querybuilder

import (
	"slices"
	"testing"

	"github.com/blevesearch/bleve/v2"
)

func TestLanguageFilters(t *testing.T) {
	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping.AddFieldMappingsAt("language", bleve.NewKeywordFieldMapping())
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := idx.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, tag := range []string{"en", "en-US", "en-GB", "nb", "nb-NO", "no", "id", "in", "unknown"} {
		if err := idx.Index(tag, map[string]string{"language": tag}); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		query string
		want  []string
	}{
		{"language:EN", []string{"en", "en-GB", "en-US"}},
		{"language:en-us", []string{"en-US"}},
		{`language:"en-us"`, []string{"en-US"}},
		{"language:en-*", []string{"en-GB", "en-US"}},
		{"language:nb", []string{"nb", "nb-NO", "no"}},
		{"language:NO", []string{"nb", "nb-NO", "no"}},
		{"language:in", []string{"id", "in"}},
		{"language:(en-us|nb)", []string{"en-US", "nb", "nb-NO", "no"}},
		{"language:en -language:en-us", []string{"en", "en-GB"}},
		{"language:unknown", []string{"unknown"}},
		{"language:zz", nil},
	} {
		t.Run(tc.query, func(t *testing.T) {
			res, err := idx.Search(bleve.NewSearchRequest(Build(tc.query)))
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, hit := range res.Hits {
				got = append(got, hit.ID)
			}
			slices.Sort(got)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("matches = %v, want %v", got, tc.want)
			}
		})
	}
}
