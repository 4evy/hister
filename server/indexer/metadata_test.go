// SPDX-License-Identifier: AGPL-3.0-or-later

package indexer

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/testutil"

	"github.com/blevesearch/bleve/v2"
)

func TestIndexMetadataPersistence(t *testing.T) {
	cfg := testutil.Config(t)
	idx, err := initializeIndexer(cfg.FullPath(""), true, true)
	if err != nil {
		t.Fatalf("initialize indexer: %v", err)
	}
	i = idx

	version, err := GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != Version {
		t.Fatalf("version = %d, want %d", version, Version)
	}
	fingerprint, err := GetAnalyzerFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	wantFingerprint := AnalyzerFingerprint(true, true)
	if fingerprint != wantFingerprint {
		t.Fatalf("fingerprint = %q, want %q", fingerprint, wantFingerprint)
	}

	if err := SetVersion(Version - 1); err != nil {
		t.Fatal(err)
	}
	if err := SetAnalyzerFingerprint("custom"); err != nil {
		t.Fatal(err)
	}
	idx.Close()

	idx, err = initializeIndexer(cfg.FullPath(""), true, true)
	if err != nil {
		t.Fatalf("reopen indexer: %v", err)
	}
	i = idx
	defer idx.Close()

	version, err = GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != Version-1 {
		t.Fatalf("reopened version = %d, want %d", version, Version-1)
	}
	fingerprint, err = GetAnalyzerFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if fingerprint != "custom" {
		t.Fatalf("reopened fingerprint = %q, want custom", fingerprint)
	}
}

func TestGetVersionRejectsInvalidInternalValue(t *testing.T) {
	cfg := testutil.Config(t)
	idx, err := initializeIndexer(cfg.FullPath(""), false, false)
	if err != nil {
		t.Fatalf("initialize indexer: %v", err)
	}
	i = idx
	defer idx.Close()

	if err := idx.indexers[defaultIndexerName].DeleteInternal([]byte(indexVersionKey)); err != nil {
		t.Fatal(err)
	}
	version, err := GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != -1 {
		t.Fatalf("missing version = %d, want -1", version)
	}

	if err := idx.indexers[defaultIndexerName].SetInternal([]byte(indexVersionKey), []byte("invalid")); err != nil {
		t.Fatal(err)
	}
	if _, err := GetVersion(); err == nil {
		t.Fatal("GetVersion accepted an invalid internal value")
	}
}

func TestLanguageIndexMetadataIsSelfContained(t *testing.T) {
	cfg := testutil.Config(t)
	idx, err := initializeIndexer(cfg.FullPath(""), true, true)
	if err != nil {
		t.Fatalf("initialize indexer: %v", err)
	}
	i = idx

	languageIndexName := indexNameForLanguage("en")
	if err := idx.addIndexer(languageIndexName, "en"); err != nil {
		idx.Close()
		t.Fatalf("add language index: %v", err)
	}
	if err := SetVersion(Version - 1); err != nil {
		idx.Close()
		t.Fatal(err)
	}
	if err := SetAnalyzerFingerprint("custom"); err != nil {
		idx.Close()
		t.Fatal(err)
	}

	for name, subIndex := range idx.indexers {
		version, err := subIndex.GetInternal([]byte(indexVersionKey))
		if err != nil {
			idx.Close()
			t.Fatalf("read version from %s: %v", name, err)
		}
		if got, want := string(version), strconv.Itoa(Version-1); got != want {
			idx.Close()
			t.Fatalf("version in %s = %q, want %q", name, got, want)
		}
		fingerprint, err := subIndex.GetInternal([]byte(analyzerConfigKey))
		if err != nil {
			idx.Close()
			t.Fatalf("read fingerprint from %s: %v", name, err)
		}
		if string(fingerprint) != "custom" {
			idx.Close()
			t.Fatalf("fingerprint in %s = %q, want custom", name, fingerprint)
		}
	}
	idx.Close()

	copiedPath := filepath.Join(t.TempDir(), languageIndexName)
	if err := os.CopyFS(copiedPath, os.DirFS(filepath.Join(cfg.FullPath(""), languageIndexName))); err != nil {
		t.Fatalf("copy language index: %v", err)
	}
	copiedIndex, err := bleve.OpenUsing(copiedPath, bleveRuntimeConfig())
	if err != nil {
		t.Fatalf("open copied language index: %v", err)
	}
	defer func() {
		if err := copiedIndex.Close(); err != nil {
			t.Errorf("close copied language index: %v", err)
		}
	}()

	version, err := copiedIndex.GetInternal([]byte(indexVersionKey))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(version), strconv.Itoa(Version-1); got != want {
		t.Fatalf("copied language index version = %q, want %q", got, want)
	}
	fingerprint, err := copiedIndex.GetInternal([]byte(analyzerConfigKey))
	if err != nil {
		t.Fatal(err)
	}
	if string(fingerprint) != "custom" {
		t.Fatalf("copied language index fingerprint = %q, want custom", fingerprint)
	}
}

func TestReindexStoresReplacementIndexMetadata(t *testing.T) {
	cfg := testutil.Config(t)
	idx, err := initializeIndexer(cfg.FullPath(""), false, false)
	if err != nil {
		t.Fatalf("initialize indexer: %v", err)
	}
	i = idx
	if err := idx.save(&document.Document{
		URL:      "https://example.com/english",
		Title:    "English document",
		Text:     "This is a deliberately long English document with enough common words for reliable language detection during the reindex operation.",
		Language: "en",
		AddCount: 1,
	}); err != nil {
		idx.Close()
		t.Fatalf("save English document: %v", err)
	}
	if err := SetVersion(Version - 1); err != nil {
		idx.Close()
		t.Fatal(err)
	}
	if err := SetAnalyzerFingerprint("old"); err != nil {
		idx.Close()
		t.Fatal(err)
	}

	if err := Reindex(
		cfg.FullPath(""),
		&config.Rules{},
		false,
		true,
		true,
		nil,
	); err != nil {
		t.Fatalf("Reindex returned an error: %v", err)
	}
	defer i.Close()

	version, err := GetVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != Version {
		t.Fatalf("version = %d, want %d", version, Version)
	}
	fingerprint, err := GetAnalyzerFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	wantFingerprint := AnalyzerFingerprint(true, true)
	if fingerprint != wantFingerprint {
		t.Fatalf("fingerprint = %q, want %q", fingerprint, wantFingerprint)
	}

	languageIndexName := indexNameForLanguage("en")
	if _, exists := i.indexers[languageIndexName]; !exists {
		t.Fatalf("replacement language index %s was not created", languageIndexName)
	}
	for name, subIndex := range i.indexers {
		storedVersion, err := subIndex.GetInternal([]byte(indexVersionKey))
		if err != nil {
			t.Fatalf("read replacement version from %s: %v", name, err)
		}
		if got, want := string(storedVersion), strconv.Itoa(Version); got != want {
			t.Fatalf("replacement version in %s = %q, want %q", name, got, want)
		}
		storedFingerprint, err := subIndex.GetInternal([]byte(analyzerConfigKey))
		if err != nil {
			t.Fatalf("read replacement fingerprint from %s: %v", name, err)
		}
		if string(storedFingerprint) != wantFingerprint {
			t.Fatalf(
				"replacement fingerprint in %s = %q, want %q",
				name,
				storedFingerprint,
				wantFingerprint,
			)
		}
	}
}
