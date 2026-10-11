package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asciimoo/hister/config"
)

func TestReloadCommand(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusBadRequest} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/hister/api/reload" || r.Header.Get("X-Access-Token") != "secret" || r.Header.Get("Origin") != "hister://" {
					t.Errorf("unexpected reload request: %s %s %v", r.Method, r.URL, r.Header)
				}
				w.WriteHeader(status)
				if status != http.StatusOK {
					_, _ = w.Write([]byte("restart required: server.address"))
				}
			}))
			defer srv.Close()
			dir := t.TempDir()
			filename := filepath.Join(dir, "client.yml")
			if err := os.WriteFile(filename, []byte("app:\n  title: Client\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("HISTER_DATA_DIR", filepath.Join(dir, "unused-data"))
			out, err := executeInspection(t, "reload", "--config", filename, "--server-url", srv.URL+"/hister", "--token", "secret")
			if calls != 1 {
				t.Fatalf("reload requests = %d, want 1; error = %v", calls, err)
			}
			if status == http.StatusOK {
				if err != nil || !strings.Contains(out, "reloaded") {
					t.Fatalf("output = %q, error = %v", out, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "server.address") {
				t.Fatalf("reload error = %v", err)
			}
			if _, err := os.Stat(filepath.Join(dir, "unused-data")); !os.IsNotExist(err) {
				t.Fatalf("reload initialized local runtime files: %v", err)
			}
		})
	}
}

func TestReloadCommandHelp(t *testing.T) {
	out, err := executeInspection(t, "reload", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Remote. Uses the configured Hister HTTP server", "--client-timeout", "--server-url", "--token"} {
		if !strings.Contains(out, want) {
			t.Errorf("reload help missing %q", want)
		}
	}
}

func TestListenConfigOverridesPreservedForReload(t *testing.T) {
	for name, value := range map[string]string{"log-level": "debug", "search-url": "https://search.example/?q={query}", "server-url": "https://hister.example/base", "token": "startup-token", "address": "127.0.0.1:5544", "public": "true"} {
		flags := rootCmd.PersistentFlags()
		if name == "address" || name == "public" {
			flags = listenCmd.Flags()
		}
		flag := flags.Lookup(name)
		oldValue, oldChanged := flag.Value.String(), flag.Changed
		t.Cleanup(func() {
			_ = flag.Value.Set(oldValue)
			flag.Changed = oldChanged
		})
		if err := flags.Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	override := listenConfigOverrides(listenCmd)
	// The closure keeps startup values even if a flag variable later changes.
	if err := rootCmd.PersistentFlags().Set("token", "later-token"); err != nil {
		t.Fatal(err)
	}
	c := config.CreateDefaultConfig()
	if err := override(c); err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.App.AccessToken != "startup-token" || !c.App.Public || c.App.LogLevel != "debug" || c.App.SearchURL != "https://search.example/?q={query}" || c.Server.Address != "127.0.0.1:5544" || c.BaseURL("") != "https://hister.example/base" {
		t.Fatalf("startup overrides not preserved: %+v", c)
	}
}
