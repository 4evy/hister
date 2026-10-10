package document

import (
	"maps"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/asciimoo/lingua-go"
)

const evidenceEnglish = "This document explains how a search engine stores web pages and helps people find useful information. Readers can search through their saved documents and return to the original pages whenever they need more details."

func TestResolveLanguagePreservesMetadata(t *testing.T) {
	detector := NewLanguageDetectorFor([]string{"en", "de"}, true)
	for _, tc := range []struct {
		name, text, title, html, override string
		disabled                          bool
	}{
		{name: "body", text: evidenceEnglish},
		{name: "title", text: "123", title: evidenceEnglish},
		{name: "hint", text: evidenceEnglish, html: `<html lang="en-US">`},
		{name: "uncertain", text: "test"},
		{name: "manual", override: "de"},
		{name: "disabled", text: evidenceEnglish, disabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, metadata := range []map[string]any{nil, {}, {"author": "Alice", "language_source": "provided by caller"}} {
				want := maps.Clone(metadata)
				d := &Document{Text: tc.text, Title: tc.title, HTML: tc.html, LanguageOverride: tc.override, Metadata: metadata}
				ld := detector
				if tc.disabled {
					ld = NewNullLanguageDetector()
				}
				d.ResolveLanguage(ld)
				if !reflect.DeepEqual(d.Metadata, want) {
					t.Errorf("metadata = %#v, want unchanged %#v", d.Metadata, want)
				}
			}
		})
	}
}

func TestDetectionAbstainsWithInsufficientEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, text    string
		first, second float64
	}{
		{"short even when confident", "hello", 1, 0},
		{"short in high accuracy", "Hi", .95, .05},
		{"near tie", evidenceEnglish, .51, .49},
		{"weak winner", evidenceEnglish, .4, .2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := NewLanguageDetectorFor(nil, false).(*langDetector)
			d.detector = fixedConfidenceDetector{values: []lingua.ConfidenceValue{
				fixedLanguageConfidence{lingua.English, tc.first},
				fixedLanguageConfidence{lingua.German, tc.second},
			}}
			if got := d.DetectLanguage(tc.text); got != UnknownLanguage {
				t.Fatalf("language = %s, want unknown", got)
			}
		})
	}
}

func TestLanguageEvidenceThresholds(t *testing.T) {
	for _, tc := range []struct {
		name, text, reason string
		first, second      float64
	}{
		{"twelve letters", "abcdefghijkl", "accepted", 1, 0},
		{"eleven letters", "abcdefghijk", "insufficient_text", 1, 0},
		{"four CJK letters", "日本語文", "accepted", 1, 0},
		{"three CJK letters", "日本語", "insufficient_text", 1, 0},
		{"confidence boundary", evidenceEnglish, "accepted", .5, .2},
		{"below confidence", evidenceEnglish, "low_confidence", .499, .2},
		{"margin boundary", evidenceEnglish, "accepted", .5, .35},
		{"below margin", evidenceEnglish, "ambiguous", .5, .3501},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := NewLanguageDetectorFor(nil, true).(*langDetector)
			d.detector = fixedConfidenceDetector{values: []lingua.ConfidenceValue{fixedLanguageConfidence{lingua.English, tc.first}, fixedLanguageConfidence{lingua.German, tc.second}}}
			result := d.DetectLanguageWithEvidence(tc.text)
			if result.Reason != tc.reason || (result.Language != "unknown") != (tc.reason == "accepted") {
				t.Fatalf("result=%+v, want %s", result, tc.reason)
			}
		})
	}
}

func TestHTMLLanguageHints(t *testing.T) {
	for _, tc := range []struct{ html, tag string }{
		{`<!doctype html><!-- header --><html lang="en-us">`, "en-US"},
		{`<html xml:lang="nb-NO">`, "nb-NO"},
		{`<html lang="en" xml:lang="de">`, "en"},
		{`<html lang="" xml:lang="de">`, ""},
		{`<body><div lang="en">`, ""},
		{`<script>"<html lang='en'>"</script>`, ""},
		{`<html lang="unknown">`, ""},
		{`<html lang="en-x-private">`, ""},
		{`<!--` + strings.Repeat("x", 64<<10) + `--><html lang="en">`, ""},
	} {
		got, ok := htmlLanguageHint(tc.html)
		if ok != (tc.tag != "") || ok && got.Tag != tc.tag {
			t.Errorf("hint=%+v, %v; want %s", got, ok, tc.tag)
		}
	}
}

func TestDocumentLanguageResolution(t *testing.T) {
	detector := NewLanguageDetectorFor([]string{"en", "de"}, true)
	for _, tc := range []struct {
		name, text, title, html, override, language string
		docType                                     DocType
		disabled                                    bool
	}{
		{name: "body", text: evidenceEnglish, language: "en"},
		{name: "title fallback", text: "123", title: evidenceEnglish, language: "en"},
		{name: "matching declaration", text: evidenceEnglish, html: `<html lang="EN-us"><body></body></html>`, language: "en-US"},
		{name: "conflicting declaration", text: evidenceEnglish, html: `<html lang="de">`, language: "en"},
		{name: "nested declaration ignored", text: evidenceEnglish, html: `<div lang="en-US">`, language: "en"},
		{name: "malformed declaration ignored", text: evidenceEnglish, html: `<html lang="../en">`, language: "en"},
		{name: "declaration cannot force short text", text: "OK", html: `<html lang="en">`, language: UnknownLanguage},
		{name: "declaration cannot defeat candidates", text: "Это русский текст", html: `<html lang="ru">`, language: UnknownLanguage},
		{name: "no filename fallback", title: evidenceEnglish, docType: Local, language: UnknownLanguage},
		{name: "manual wins", text: evidenceEnglish, override: "de", language: "de"},
		{name: "disabled ignores title and hint", title: evidenceEnglish, html: `<html lang="en-US">`, disabled: true, language: UnknownLanguage},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &Document{Text: tc.text, Title: tc.title, HTML: tc.html, LanguageOverride: tc.override, Type: tc.docType}
			ld := detector
			if tc.disabled {
				ld = NewNullLanguageDetector()
			}
			d.ResolveLanguage(ld)
			if d.Language != tc.language {
				t.Errorf("language = %s, want %s", d.Language, tc.language)
			}
		})
	}
}

func TestAmbiguousBodyDoesNotFallBackToTitle(t *testing.T) {
	d := NewLanguageDetectorFor(nil, true).(*langDetector)
	d.detector = fixedConfidenceDetector{values: []lingua.ConfidenceValue{
		fixedLanguageConfidence{lingua.English, .51}, fixedLanguageConfidence{lingua.German, .49},
	}}
	doc := &Document{Text: evidenceEnglish, Title: evidenceEnglish}
	doc.ResolveLanguage(d)
	if doc.Language != UnknownLanguage {
		t.Fatalf("language = %s, want unknown", doc.Language)
	}
}

func TestLanguageEvidenceReportsScoresWhenAbstaining(t *testing.T) {
	d := NewLanguageDetectorFor(nil, true).(*langDetector)
	d.detector = fixedConfidenceDetector{values: []lingua.ConfidenceValue{
		fixedLanguageConfidence{lingua.English, .51}, fixedLanguageConfidence{lingua.German, .49},
	}}
	result := d.DetectLanguageWithEvidence(evidenceEnglish)
	if result.Language != UnknownLanguage || result.Candidate != "en" || result.RunnerUp != "de" || result.Confidence != .51 || result.RunnerUpConfidence != .49 {
		t.Fatalf("unexpected detection result: %+v", result)
	}
	if math.Abs(result.Margin-.02) > 1e-10 {
		t.Fatalf("invalid margin: %+v", result)
	}
}
