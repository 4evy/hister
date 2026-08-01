package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/asciimoo/hister/server/testutil"
)

type statusCountingWriter struct {
	http.ResponseWriter
	statusCodes []int
}

func (w *statusCountingWriter) WriteHeader(statusCode int) {
	w.statusCodes = append(w.statusCodes, statusCode)
	w.ResponseWriter.WriteHeader(statusCode)
}

func TestAddFormAccessTokenDoesNotAuthenticate(t *testing.T) {
	cfg := testutil.Config(t)
	cfg.App.AccessToken = "secret"
	sessionStore = newSessionStore([]byte(strings.Repeat("x", 32)), sessionMaxAge)
	called := false
	h := endpointHandler(func(c *webContext) {
		called = true
		c.Response.WriteHeader(http.StatusNoContent)
	})
	h = withCSRF(h)
	h = withTokenAuth(h)
	handler := http.HandlerFunc(createHandler(cfg, h))
	body := url.Values{
		"access_token": {"secret"},
		"url":          {"https://example.com"},
	}.Encode()

	rec := testutil.ServeHTTP(
		t,
		handler,
		http.MethodPost,
		"/api/add",
		strings.NewReader(body),
		map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Origin":       "https://unrelated.example",
		},
	)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /api/add status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if called {
		t.Fatal("POST /api/add with a form access token reached the protected handler")
	}
}

func TestServeAddFormWritesStatusOnce(t *testing.T) {
	cfg := testutil.Config(t)
	cfg.Server.BaseURL = "http://127.0.0.1:4433"
	body := url.Values{
		"url": {"http://127.0.0.1:4433/already-local"},
	}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/api/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	writer := &statusCountingWriter{ResponseWriter: rec}

	serveAdd(&webContext{
		Request:  req,
		Response: writer,
		Config:   cfg,
	})

	if len(writer.statusCodes) != 1 {
		t.Fatalf("WriteHeader calls = %v, want one call", writer.statusCodes)
	}
	if writer.statusCodes[0] != http.StatusNotAcceptable {
		t.Fatalf("status = %d, want %d", writer.statusCodes[0], http.StatusNotAcceptable)
	}
}
