package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReloadPreservesRuntimeStateAndAppliesOverrides(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "config.yml")
	write := func(title string) {
		t.Helper()
		content := fmt.Sprintf("app:\n  directory: %q\n  title: %s\n", dir, title)
		if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("Original")
	c, err := Load(filename)
	if err != nil {
		t.Fatal(err)
	}
	write("Reloaded")
	t.Setenv("HISTER__APP__SUBTITLE", "From server environment")
	if err := os.WriteFile(filepath.Join(dir, "tui.yaml"), []byte("dark_theme: changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, secretKeyFilename), []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	next, err := c.Reload(func(next *Config) error {
		next.App.SearchURL = "https://startup.example/?q={query}"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if next.App.Title != "Reloaded" || next.App.Subtitle != "From server environment" || next.App.SearchURL != "https://startup.example/?q={query}" {
		t.Fatalf("incorrect reload: %+v", next.App)
	}
	if c.App.Title != "Original" || next.TUI != c.TUI || !bytes.Equal(next.SecretKey(), c.SecretKey()) {
		t.Fatal("reload mutated original configuration or replaced startup runtime state")
	}
}

func TestReloadRequiresOriginalFile(t *testing.T) {
	dir := t.TempDir()
	c := CreateDefaultConfig()
	c.fname = filepath.Join(dir, "missing.yml")
	c.App.Directory = dir
	if _, err := c.Reload(); err == nil {
		t.Fatal("reload silently fell back from the original config file")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed reload created files: %v, %v", entries, err)
	}
}
