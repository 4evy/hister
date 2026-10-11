package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestInitLogUsesReloadableOutput(t *testing.T) {
	dir := t.TempDir()
	previous, level := cfg, zerolog.GlobalLevel()
	t.Cleanup(func() {
		_ = logging.Default.Close()
		cfg = previous
		zerolog.SetGlobalLevel(level)
	})
	cfg = config.CreateDefaultConfig()
	cfg.App.LogFile = filepath.Join(dir, "startup.log")
	cfg.App.LogFormat = "json"
	initLog()
	derived := log.With().Str("component", "existing").Logger()
	derived.Info().Msg("startup-output")
	updated := cfg.App
	updated.LogFile = filepath.Join(dir, "reloaded.log")
	change, err := logging.Default.Prepare(updated)
	if err != nil {
		t.Fatal(err)
	}
	defer change.Close()
	if err := change.Apply(); err != nil {
		t.Fatal(err)
	}
	derived.Info().Msg("reloaded-output")
	for path, want := range map[string]string{cfg.App.LogFile: "startup-output", updated.LogFile: "reloaded-output"} {
		raw, err := os.ReadFile(path)
		if err != nil || !json.Valid(raw) || !strings.Contains(string(raw), want) {
			t.Fatalf("incorrect log output in %s: %s (%v)", path, raw, err)
		}
	}
}
