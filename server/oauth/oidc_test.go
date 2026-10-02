package oauth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOIDCIdentityClaim(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		claim   string
		body    string
		want    UserInfoResponse
		wantErr string
	}{
		{
			name: "default email", body: `{"email":"a@example.com","sub":"s1","preferred_username":"alice"}`,
			want: UserInfoResponse{UID: "oidc-a@example.com", Email: "a@example.com", Username: "alice"},
		},
		{
			name: "explicit email preserves identity", claim: "email", body: `{"email":"a@example.com","sub":"s1"}`,
			want: UserInfoResponse{UID: "oidc-a@example.com", Email: "a@example.com"},
		},
		{
			name: "subject", claim: "sub", body: `{"email":"a@example.com","sub":"s1","preferred_username":"alice"}`,
			want: UserInfoResponse{UID: "oidc:sub:s1", Email: "a@example.com", Username: "alice"},
		},
		{name: "subject alone", claim: "sub", body: `{"sub":"s1"}`, want: UserInfoResponse{UID: "oidc:sub:s1"}},
		{name: "null optional claims", claim: "sub", body: `{"sub":"s1","email":null,"preferred_username":null}`, want: UserInfoResponse{UID: "oidc:sub:s1"}},
		{name: "custom claim", claim: "account_id", body: `{"account_id":"s1"}`, want: UserInfoResponse{UID: "oidc:account_id:s1"}},
		{name: "literal claim name", claim: "account.id", body: `{"account.id":"s1","account":{"id":"other"}}`, want: UserInfoResponse{UID: "oidc:account.id:s1"}},
		{name: "URI claim", claim: "https://example.com/subject", body: `{"https://example.com/subject":"s1"}`, want: UserInfoResponse{UID: "oidc:https%3A%2F%2Fexample.com%2Fsubject:s1"}},
		{name: "escape claim separator", claim: "custom:sub", body: `{"custom:sub":"s1"}`, want: UserInfoResponse{UID: "oidc:custom%3Asub:s1"}},
		{name: "preserve value separator", claim: "custom", body: `{"custom":"sub:s1"}`, want: UserInfoResponse{UID: "oidc:custom:sub:s1"}},
		{name: "preserve exact value", claim: "sub", body: `{"sub":" S1 "}`, want: UserInfoResponse{UID: "oidc:sub: S1 "}},
		{name: "missing email by default", body: `{"sub":"s1"}`, wantErr: `"email"`},
		{name: "missing subject never falls back", claim: "sub", body: `{"email":"a@example.com"}`, wantErr: `"sub"`},
		{name: "case sensitive claim", claim: "sub", body: `{"SUB":"s1"}`, wantErr: `"sub"`},
		{name: "empty subject", claim: "sub", body: `{"sub":""}`, wantErr: `"sub"`},
		{name: "blank subject", claim: "sub", body: `{"sub":" \t\n\u00a0"}`, wantErr: `"sub"`},
		{name: "null subject", claim: "sub", body: `{"sub":null}`, wantErr: `"sub"`},
		{name: "numeric subject", claim: "sub", body: `{"sub":123}`, wantErr: `"sub"`},
		{name: "boolean subject", claim: "sub", body: `{"sub":true}`, wantErr: `"sub"`},
		{name: "array subject", claim: "sub", body: `{"sub":["s1"]}`, wantErr: `"sub"`},
		{name: "object subject", claim: "sub", body: `{"sub":{"id":"s1"}}`, wantErr: `"sub"`},
		{name: "malformed JSON", claim: "sub", body: `{"sub":`, wantErr: "failed to parse UserInfo"},
		{name: "array response", claim: "sub", body: `[]`, wantErr: "failed to parse UserInfo"},
		{name: "null response", claim: "sub", body: `null`, wantErr: `"sub"`},
		{name: "invalid email type", claim: "sub", body: `{"sub":"s1","email":123}`, wantErr: `"email"`},
		{name: "invalid username type", claim: "sub", body: `{"sub":"s1","preferred_username":[]}`, wantErr: `"preferred_username"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			provider, ok := NewProvider("oidc", ProviderConfig{UserInfoURL: "https://provider.example/userinfo", IdentityClaim: tt.claim})
			if !ok {
				t.Fatal("missing OIDC provider")
			}
			provider.(*OIDCOAuth).Client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://provider.example/userinfo" || req.Header.Get("Authorization") != "Bearer token" {
					t.Error("invalid UserInfo request")
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(tt.body))}, nil
			})}
			info, err := provider.GetUserInfo(context.Background(), TokenResponse(`{"access_token":"token"}`))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) || info != nil {
					t.Fatalf("GetUserInfo = %+v, %v; want error containing %q", info, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if *info != tt.want {
				t.Fatalf("GetUserInfo = %+v, want %+v", info, tt.want)
			}
		})
	}
}

func TestOIDCDiscoveryPreservesIdentityClaim(t *testing.T) {
	t.Parallel()
	provider := &OIDCOAuth{
		IdentityClaim: "sub",
		Client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{
				"authorization_endpoint":"https://provider.example/auth",
				"token_endpoint":"https://provider.example/token",
				"userinfo_endpoint":"https://provider.example/userinfo",
				"scopes_supported":["openid"], "response_types_supported":["code"],
				"grant_types_supported":["authorization_code"],
				"IdentityClaim":"email", "identity_claim":"email"
			}`))}, nil
		})},
	}
	if err := provider.Prepare(context.Background(), NewPrepareRequest("https://provider.example/discovery")); err != nil {
		t.Fatal(err)
	}
	if provider.IdentityClaim != "sub" {
		t.Fatalf("discovery changed identity claim to %q", provider.IdentityClaim)
	}
}
