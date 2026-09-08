// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/model"
	"github.com/asciimoo/hister/server/testutil"
)

func newOAuthTestServer(t *testing.T) (*config.Config, http.Handler) {
	t.Helper()
	cfg := testutil.InitModel(t)
	cfg.App.UserHandling = true
	cfg.Server.OAuth = map[string]*config.OAuthEntry{
		"oidc":   {ClientID: "hister", ClientSecret: "secret", AuthURL: "https://provider.example/auth"},
		"google": {ClientID: "google-client", ClientSecret: "google-secret"},
	}
	if err := cfg.UpdateBaseURL("http://hister.example/hister"); err != nil {
		t.Fatal(err)
	}
	previousStore := sessionStore
	t.Cleanup(func() { sessionStore = previousStore })
	sessionStore = newSessionStore([]byte(strings.Repeat("x", 32)), cfg.BaseURL(""), sessionMaxAge)
	return cfg, registerEndpoints(cfg, nil)
}

func startOAuthLogin(t *testing.T, handler http.Handler, cookie *http.Cookie) (*http.Cookie, url.Values) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/hister/api/oauth?provider=oidc", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("OAuth redirect status = %d: %s", rec.Code, rec.Body.String())
	}
	redirect, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	return responseSessionCookie(t, rec), redirect.Query()
}

func oauthCallback(handler http.Handler, cookie *http.Cookie, query url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/hister/api/oauth/callback?"+query.Encode(), nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func oauthSessionValues(t *testing.T, cookie *http.Cookie) map[any]any {
	t.Helper()
	record, err := model.GetWebSession(sessionTokenHash(cookie.Value))
	if err != nil {
		t.Fatal(err)
	}
	values := make(map[any]any)
	if err := decodeSessionValues(record.Data, &values); err != nil {
		t.Fatal(err)
	}
	return values
}

func assertOAuthFlowConsumed(t *testing.T, cookie *http.Cookie) {
	t.Helper()
	values := oauthSessionValues(t, cookie)
	for _, key := range []string{oauthStateKey, oauthProviderKey, oauthVerifierKey} {
		if _, ok := values[key]; ok {
			t.Errorf("session still contains %s", key)
		}
	}
}

func TestOAuthPKCEFlow(t *testing.T) {
	for _, mode := range []string{"manual", "discovery", "disabled"} {
		t.Run(mode, func(t *testing.T) {
			cfg, handler := newOAuthTestServer(t)
			cfg.Server.OAuthOnly = true
			entry := cfg.Server.OAuth["oidc"]
			entry.DisablePKCE = mode == "disabled"
			var cookie *http.Cookie
			var challenge string
			var tokenCalls atomic.Int32
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/discovery":
					_ = json.NewEncoder(w).Encode(map[string]any{
						"authorization_endpoint": entry.AuthURL, "token_endpoint": entry.TokenURL,
						"userinfo_endpoint": entry.UserInfoURL, "scopes_supported": []string{"openid"},
						"response_types_supported": []string{"code"}, "grant_types_supported": []string{"authorization_code"},
						// Deliberately omit PKCE metadata. It is not required to use S256.
					})
				case "/token":
					tokenCalls.Add(1)
					record, err := model.GetWebSession(sessionTokenHash(cookie.Value))
					values := make(map[any]any)
					if err == nil {
						err = decodeSessionValues(record.Data, &values)
					}
					if err != nil {
						t.Errorf("read session during token exchange: %v", err)
						http.Error(w, "session failure", http.StatusInternalServerError)
						return
					}
					for _, key := range []string{oauthStateKey, oauthProviderKey, oauthVerifierKey} {
						if _, ok := values[key]; ok {
							t.Errorf("token exchange started before consuming %s", key)
						}
					}
					if err := r.ParseForm(); err != nil {
						t.Error(err)
						http.Error(w, "invalid form", http.StatusBadRequest)
						return
					}
					verifier := r.PostForm.Get("code_verifier")
					digest := sha256.Sum256([]byte(verifier))
					if mode != "disabled" && (len(verifier) != 43 || base64.RawURLEncoding.EncodeToString(digest[:]) != challenge) {
						t.Error("token verifier does not match the authorization challenge")
						http.Error(w, "invalid verifier", http.StatusBadRequest)
						return
					}
					if mode == "disabled" && r.PostForm.Has("code_verifier") {
						t.Error("disabled PKCE sent a verifier")
					}
					for key, want := range map[string]string{
						"client_id": "hister", "client_secret": "secret", "code": "authorization-code",
						"redirect_uri": cfg.BaseURL("/api/oauth/callback") + "?provider=oidc",
					} {
						if got := r.PostForm.Get(key); got != want {
							t.Errorf("token %s = %q, want %q", key, got, want)
						}
					}
					_, _ = fmt.Fprint(w, `{"access_token":"token"}`)
				case "/userinfo":
					if r.Header.Get("Authorization") != "Bearer token" {
						t.Error("missing userinfo authorization")
					}
					_, _ = fmt.Fprint(w, `{"email":"alice@example.com","preferred_username":"alice"}`)
				default:
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(provider.Close)
			entry.AuthURL = provider.URL + "/authorize"
			entry.TokenURL = provider.URL + "/token"
			entry.UserInfoURL = provider.URL + "/userinfo"
			if mode == "discovery" {
				entry.ConfigurationURL = provider.URL + "/discovery"
			}

			// A second login must reuse the account and replace the existing session.
			for range 2 {
				var query url.Values
				cookie, query = startOAuthLogin(t, handler, cookie)
				challenge = query.Get("code_challenge")
				values := oauthSessionValues(t, cookie)
				if values[oauthProviderKey] != "oidc" || values[oauthStateKey] != query.Get("state") || query.Get("state") == "" {
					t.Fatal("authorization request is not bound to the session")
				}
				if mode == "disabled" {
					if query.Has("code_challenge") || query.Has("code_challenge_method") {
						t.Fatal("disabled PKCE sent a challenge")
					}
				} else if challenge == "" || query.Get("code_challenge_method") != "S256" || values[oauthVerifierKey] == "" {
					t.Fatal("missing PKCE challenge or verifier")
				}
				if query.Has("code_verifier") || strings.Contains(cookie.String(), fmt.Sprint(values[oauthVerifierKey])) && mode != "disabled" {
					t.Fatal("verifier exposed to browser")
				}
				callback := url.Values{"provider": {"oidc"}, "state": {query.Get("state")}, "code": {"authorization-code"}}
				rec := oauthCallback(handler, cookie, callback)
				if rec.Code != http.StatusFound || rec.Header().Get("Location") != cfg.BaseURL("/") {
					t.Fatalf("callback status = %d, body = %s", rec.Code, rec.Body.String())
				}
				oldCookie := cookie
				cookie = responseSessionCookie(t, rec)
				if oldCookie.Value == cookie.Value {
					t.Fatal("login did not rotate the session")
				}
				assertOAuthFlowConsumed(t, cookie)
				user, err := model.GetUserByOAuthID("oidc-alice@example.com")
				if err != nil {
					t.Fatal(err)
				}
				if oauthSessionValues(t, cookie)["user_id"] != user.ID {
					t.Fatal("session does not authenticate the OAuth user")
				}
				if got := oauthCallback(handler, oldCookie, callback).Code; got != http.StatusBadRequest {
					t.Fatalf("replayed callback status = %d", got)
				}
			}
			if tokenCalls.Load() != 2 {
				t.Fatalf("token requests = %d, want 2", tokenCalls.Load())
			}
			var count int64
			if err := model.DB.Model(&model.User{}).Count(&count).Error; err != nil || count != 1 {
				t.Fatalf("users = %d, error = %v", count, err)
			}
		})
	}
}

func TestOAuthCallbackFailures(t *testing.T) {
	for _, failure := range []string{"state", "provider", "cookie", "legacy session", "missing verifier", "empty verifier", "denied", "missing code", "token", "userinfo"} {
		t.Run(failure, func(t *testing.T) {
			cfg, handler := newOAuthTestServer(t)
			var tokenCalls atomic.Int32
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/token" {
					tokenCalls.Add(1)
					if failure == "userinfo" {
						_, _ = fmt.Fprint(w, `{"access_token":"token"}`)
						return
					}
				}
				http.Error(w, "provider failure", http.StatusBadGateway)
			}))
			t.Cleanup(provider.Close)
			cfg.Server.OAuth["oidc"].TokenURL = provider.URL + "/token"
			cfg.Server.OAuth["oidc"].UserInfoURL = provider.URL + "/userinfo"
			cookie, query := startOAuthLogin(t, handler, nil)
			callback := url.Values{"provider": {"oidc"}, "state": {query.Get("state")}, "code": {"code"}}
			wantStatus := http.StatusBadRequest
			wantCalls := int32(0)
			consumed := true
			switch failure {
			case "state":
				callback.Set("state", "incorrect")
				consumed = false
			case "provider":
				callback.Set("provider", "google")
				consumed = false
			case "cookie":
				cookie = nil
				consumed = false
			case "legacy session", "missing verifier", "empty verifier":
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(cookie)
				session, err := sessionStore.Get(req, storeName)
				if err != nil {
					t.Fatal(err)
				}
				delete(session.Values, oauthVerifierKey)
				switch failure {
				case "empty verifier":
					session.Values[oauthVerifierKey] = ""
				case "legacy session":
					delete(session.Values, oauthProviderKey)
					consumed = false
				}
				if err := session.Save(req, httptest.NewRecorder()); err != nil {
					t.Fatal(err)
				}
			case "denied":
				callback.Set("error", "access_denied")
			case "missing code":
				callback.Del("code")
			case "token", "userinfo":
				wantStatus = http.StatusInternalServerError
				wantCalls = 1
			}
			if rec := oauthCallback(handler, cookie, callback); rec.Code != wantStatus {
				t.Fatalf("callback status = %d, want %d: %s", rec.Code, wantStatus, rec.Body.String())
			}
			if consumed {
				assertOAuthFlowConsumed(t, cookie)
				if got := oauthCallback(handler, cookie, callback).Code; got != http.StatusBadRequest {
					t.Fatalf("replayed callback status = %d", got)
				}
			}
			if got := tokenCalls.Load(); got != wantCalls {
				t.Fatalf("token requests = %d, want %d", got, wantCalls)
			}
		})
	}
}

func TestOAuthNewLoginReplacesPendingFlow(t *testing.T) {
	_, handler := newOAuthTestServer(t)
	cookie, first := startOAuthLogin(t, handler, nil)
	cookie, second := startOAuthLogin(t, handler, cookie)
	if first.Get("state") == second.Get("state") || first.Get("code_challenge") == second.Get("code_challenge") {
		t.Fatal("new login reused state or verifier")
	}
	callback := url.Values{"provider": {"oidc"}, "state": {first.Get("state")}, "code": {"code"}}
	if got := oauthCallback(handler, cookie, callback).Code; got != http.StatusBadRequest {
		t.Fatalf("replaced flow status = %d", got)
	}
	if oauthSessionValues(t, cookie)[oauthStateKey] != second.Get("state") {
		t.Fatal("invalid callback consumed the current flow")
	}
}
