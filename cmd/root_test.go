// SPDX-FileContributor: 4evy <git@evy.pink>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"testing"

	"github.com/asciimoo/hister/server/indexer"
)

func TestInitialAnalyzerFingerprint(t *testing.T) {
	tests := []struct {
		name            string
		indexerVersion  int
		detectLanguages bool
		keepStopwords   bool
		want            string
	}{
		{
			name:            "fresh index uses active configuration",
			indexerVersion:  -1,
			detectLanguages: true,
			keepStopwords:   true,
			want:            indexer.AnalyzerFingerprint(true, true),
		},
		{
			name:            "upgraded index with defaults matches active configuration",
			indexerVersion:  indexer.Version,
			detectLanguages: true,
			keepStopwords:   false,
			want:            indexer.AnalyzerFingerprint(true, false),
		},
		{
			name:            "upgraded index uses legacy configuration",
			indexerVersion:  indexer.Version,
			detectLanguages: true,
			keepStopwords:   true,
			want:            indexer.AnalyzerFingerprint(true, false),
		},
		{
			name:            "upgraded index retains disabled language detection",
			indexerVersion:  indexer.Version,
			detectLanguages: false,
			keepStopwords:   true,
			want:            indexer.AnalyzerFingerprint(false, false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := initialAnalyzerFingerprint(tt.indexerVersion, tt.detectLanguages, tt.keepStopwords)
			if got != tt.want {
				t.Fatalf("initialAnalyzerFingerprint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmbeddingConfigWarning(t *testing.T) {
	tests := []struct {
		name   string
		stored string
		active string
		want   string
	}{
		{
			name:   "matching configuration",
			stored: "same",
			active: "same",
		},
		{
			name:   "semantic search disabled",
			stored: "stored",
		},
		{
			name:   "legacy indexed configuration",
			active: "active",
		},
		{
			name:   "changed configuration",
			stored: "stored",
			active: "active",
			want:   "The semantic search embedding configuration differs from the indexed configuration. Run `hister reindex` to update your embeddings.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := embeddingConfigWarning(tt.stored, tt.active); got != tt.want {
				t.Fatalf("embeddingConfigWarning() = %q, want %q", got, tt.want)
			}
		})
	}
}
