package querybuilder

import (
	"fmt"
	"testing"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

func BenchmarkURLRegexp(b *testing.B) {
	m := bleve.NewIndexMapping()
	m.DefaultMapping.AddFieldMappingsAt("url", bleve.NewKeywordFieldMapping())
	idx, err := bleve.NewUsing("", m, bleve.Config.DefaultIndexType, bleve.Config.DefaultMemKVStore, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = idx.Close() })
	batch := idx.NewBatch()
	for i := range 10000 {
		domain := "common.test"
		if i%400 == 0 {
			domain = "rare.test"
		}
		if err := batch.Index(fmt.Sprint(i), map[string]string{"url": fmt.Sprintf("https://%s/entry/%d", domain, i)}); err != nil {
			b.Fatal(err)
		}
	}
	if err := idx.Batch(batch); err != nil {
		b.Fatal(err)
	}
	for _, tc := range []struct{ name, pattern string }{
		{"anchored", `^https://rare\.test/entry/\d+$`},
		{"unanchored", `rare\.test/entry/`},
		{"no_matches", `^https://missing\.test/`},
		{"broad", `^https://common\.test/entry/`},
	} {
		b.Run(tc.name, func(b *testing.B) {
			optimized, err := BuildValidated(`url_re:"` + tc.pattern + `"`)
			if err != nil {
				b.Fatal(err)
			}
			for _, implementation := range []struct {
				name string
				q    query.Query
			}{
				{"document_scan", query.NewBooleanQuery([]query.Query{legacyURLRegexp(tc.pattern)}, nil, nil)},
				{"indexed_terms", optimized},
			} {
				b.Run(implementation.name, func(b *testing.B) {
					req := bleve.NewSearchRequestOptions(implementation.q, 20, 0, false)
					b.ReportAllocs()
					for b.Loop() {
						if _, err := idx.Search(req); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}
