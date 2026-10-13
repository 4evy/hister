package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSnippetLengthConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		env  string
		want int
	}{
		{"default", "", "", 300},
		{"yaml", "app:\n  snippet_length: 800\n", "", 800},
		{"environment", "app:\n  snippet_length: 800\n", "1000", 1000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.env != "" {
				t.Setenv("HISTER__APP__SNIPPET_LENGTH", tc.env)
			}
			cfg, err := parseConfig([]byte(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			data, err := yaml.Marshal(cfg)
			if err != nil {
				t.Fatal(err)
			}
			var generated struct {
				App     map[string]any `yaml:"app"`
				Indexer map[string]any `yaml:"indexer"`
			}
			if err := yaml.Unmarshal(data, &generated); err != nil {
				t.Fatal(err)
			}
			if got := generated.App["snippet_length"]; got != tc.want {
				t.Errorf("generated snippet_length = %v, want %d", got, tc.want)
			}
			if _, ok := generated.Indexer["snippet_length"]; ok {
				t.Error("generated config still includes indexer.snippet_length")
			}
		})
	}
}

func TestSnippetLengthValidation(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		t.Run(value, func(t *testing.T) {
			cfg, err := parseConfig([]byte("app:\n  snippet_length: " + value + "\n"))
			if err == nil {
				err = cfg.Validate()
			}
			if err == nil || !strings.Contains(err.Error(), "app.snippet_length") {
				t.Fatalf("invalid snippet length error = %v", err)
			}
		})
	}
}
