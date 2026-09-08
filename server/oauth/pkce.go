// SPDX-License-Identifier: AGPL-3.0-or-later

package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
)

// NewCodeVerifier creates a fresh PKCE verifier with 256 bits of entropy.
// The verifier must be kept private until the authorization code is exchanged.
func NewCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// WithPKCE adds an S256 challenge derived from the verifier to the request.
// An empty verifier leaves PKCE disabled.
func (r *RedirectURIRequest) WithPKCE(verifier string) *RedirectURIRequest {
	r.codeChallenge = ""
	if verifier != "" {
		digest := sha256.Sum256([]byte(verifier))
		r.codeChallenge = base64.RawURLEncoding.EncodeToString(digest[:])
	}
	return r
}

// WithPKCE adds the verifier used for the authorization request to the token exchange.
// An empty verifier leaves PKCE disabled.
func (r *TokenRequest) WithPKCE(verifier string) *TokenRequest {
	r.codeVerifier = verifier
	return r
}

func (r *RedirectURIRequest) addPKCE(params url.Values) {
	if r.codeChallenge != "" {
		params.Set("code_challenge", r.codeChallenge)
		params.Set("code_challenge_method", "S256")
	}
}

func (r *TokenRequest) addPKCE(params url.Values) {
	if r.codeVerifier != "" {
		params.Set("code_verifier", r.codeVerifier)
	}
}
