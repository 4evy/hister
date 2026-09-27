// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/asciimoo/hister/config"
)

func TestHTTPClientRequestSettings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "custom agent" || r.Header.Get("X-Test") != "import" {
			t.Errorf("request headers = %v", r.Header)
		}
		cookie, err := r.Cookie("session")
		if r.URL.Path == "/private/page" {
			if err != nil || cookie.Value != "secret" {
				t.Errorf("missing configured cookie: %v", err)
			}
		} else if err == nil {
			t.Error("cookie sent outside its configured path")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewHTTPClient(&config.CrawlerConfig{
		Timeout:   2,
		UserAgent: "default agent",
		Headers:   map[string]string{"User-Agent": "custom agent", "X-Test": "import"},
		Cookies:   []config.CrawlerCookie{{Name: "session", Value: "secret", Domain: serverURL.Hostname(), Path: "/private"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.CloseIdleConnections()
	if client.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", client.Timeout)
	}
	for _, path := range []string{"/private/page", "/public"} {
		resp, err := client.Get(t.Context(), server.URL+path)
		if err != nil {
			t.Fatal(err)
		}
		_, readErr := io.Copy(io.Discard, resp.Body)
		closeErr := resp.Body.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.Get(ctx, server.URL); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled request returned %v", err)
	}
}
