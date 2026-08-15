package extractor

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/asciimoo/hister/server/types"
)

const liveManifestPath = "live_cases.yaml"

type liveManifest struct {
	Version int        `json:"version" yaml:"version"`
	Cases   []liveCase `json:"cases"   yaml:"cases"`
}

type liveCase struct {
	Name           string         `json:"name"                      yaml:"name"`
	URL            string         `json:"url"                       yaml:"url"`
	Backend        string         `json:"backend"                   yaml:"backend"`
	BackendOptions map[string]any `json:"backend_options,omitempty" yaml:"backend_options,omitempty"`
	Extractor      string         `json:"extractor"                 yaml:"extractor"`
	Timeout        int            `json:"timeout,omitempty"         yaml:"timeout,omitempty"`
	Match          *bool          `json:"match,omitempty"           yaml:"match,omitempty"`
	ExtractState   string         `json:"extract_state,omitempty"   yaml:"extract_state,omitempty"`
	RunChain       *bool          `json:"run_chain,omitempty"       yaml:"run_chain,omitempty"`
	Expect         liveExpect     `json:"expect,omitempty"          yaml:"expect,omitempty"`
}

type liveExpect struct {
	FinalURLContains   string             `json:"final_url_contains,omitempty" yaml:"final_url_contains,omitempty"`
	TitleContains      []string           `json:"title_contains,omitempty"     yaml:"title_contains,omitempty"`
	TextContains       []string           `json:"text_contains,omitempty"      yaml:"text_contains,omitempty"`
	TextNotContains    []string           `json:"text_not_contains,omitempty"  yaml:"text_not_contains,omitempty"`
	MinTextLength      int                `json:"min_text_length,omitempty"     yaml:"min_text_length,omitempty"`
	Metadata           map[string]any     `json:"metadata,omitempty"           yaml:"metadata,omitempty"`
	MetadataMinimums   map[string]float64 `json:"metadata_minimums,omitempty"  yaml:"metadata_minimums,omitempty"`
	AbsentMetadata     []string           `json:"absent_metadata,omitempty"    yaml:"absent_metadata,omitempty"`
	PreviewState       string             `json:"preview_state,omitempty"      yaml:"preview_state,omitempty"`
	PreviewContains    []string           `json:"preview_contains,omitempty"   yaml:"preview_contains,omitempty"`
	PreviewNotContains []string           `json:"preview_not_contains,omitempty" yaml:"preview_not_contains,omitempty"`
	MinPreviewLength   int                `json:"min_preview_length,omitempty"  yaml:"min_preview_length,omitempty"`
}

func TestLiveExtractorManifest(t *testing.T) {
	manifest, err := loadLiveManifest()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateLiveManifest(manifest); err != nil {
		t.Fatal(err)
	}
}

func loadLiveManifest() (*liveManifest, error) {
	data, err := os.ReadFile(liveManifestPath)
	if err != nil {
		return nil, fmt.Errorf("read live extractor manifest: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var manifest liveManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode live extractor manifest: %w", err)
	}
	return &manifest, nil
}

func validateLiveManifest(manifest *liveManifest) error {
	if manifest.Version != 1 {
		return fmt.Errorf("live extractor manifest version is %d, want 1", manifest.Version)
	}
	if len(manifest.Cases) == 0 {
		return fmt.Errorf("live extractor manifest has no cases")
	}
	seen := make(map[string]struct{}, len(manifest.Cases))
	for index, testCase := range manifest.Cases {
		prefix := fmt.Sprintf("live extractor case %d", index+1)
		if testCase.Name == "" {
			return fmt.Errorf("%s has no name", prefix)
		}
		if _, exists := seen[testCase.Name]; exists {
			return fmt.Errorf("duplicate live extractor case name %q", testCase.Name)
		}
		seen[testCase.Name] = struct{}{}
		parsedURL, err := url.Parse(testCase.URL)
		if err != nil || parsedURL.Hostname() == "" ||
			(parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return fmt.Errorf("live extractor case %q has invalid URL %q", testCase.Name, testCase.URL)
		}
		switch testCase.Backend {
		case "", "http", "chromedp", "bidi":
		default:
			return fmt.Errorf("live extractor case %q has unknown backend %q", testCase.Name, testCase.Backend)
		}
		if liveExtractorByName(testCase.Extractor) == nil {
			return fmt.Errorf("live extractor case %q names unknown extractor %q", testCase.Name, testCase.Extractor)
		}
		if _, err := parseLiveState(testCase.ExtractState, types.ExtractorStop); err != nil {
			return fmt.Errorf("live extractor case %q: %w", testCase.Name, err)
		}
		if _, err := parseLiveState(testCase.Expect.PreviewState, types.ExtractorStop); err != nil {
			return fmt.Errorf("live extractor case %q: %w", testCase.Name, err)
		}
	}
	return nil
}

func liveExtractorByName(name string) Extractor {
	for _, candidate := range extractors {
		if strings.EqualFold(candidate.Name(), name) {
			return candidate
		}
	}
	return nil
}

func parseLiveState(value string, defaultState types.ExtractorState) (types.ExtractorState, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return defaultState, nil
	case "stop":
		return types.ExtractorStop, nil
	case "continue":
		return types.ExtractorContinue, nil
	case "abort":
		return types.ExtractorAbort, nil
	default:
		return 0, fmt.Errorf("unknown extractor state %q", value)
	}
}

func liveBool(value *bool, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}
	return *value
}
