package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOAuthPKCEConfig(t *testing.T) {
	const raw = "server:\n  oauth:\n    oidc:\n      client_id: hister\n      client_secret: secret\n      auth_url: https://provider.example/auth\n"
	for _, setting := range []string{"", "      disable_pkce: false\n", "      disable_pkce: true\n"} {
		cfg, err := parseConfig([]byte(raw + setting))
		if err != nil {
			t.Fatal(err)
		}
		wantDisabled := setting == "      disable_pkce: true\n"
		if cfg.Server.OAuth["oidc"].DisablePKCE != wantDisabled {
			t.Fatalf("disable_pkce for %q = %v", setting, cfg.Server.OAuth["oidc"].DisablePKCE)
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		reloaded, err := parseConfig(data)
		if err != nil {
			t.Fatal(err)
		}
		if reloaded.Server.OAuth["oidc"].DisablePKCE != wantDisabled {
			t.Fatal("disable_pkce changed after saving config")
		}
	}
	for _, value := range []string{"true", "false"} {
		t.Setenv("HISTER__SERVER__OAUTH__OIDC__DISABLE_PKCE", value)
		cfg, err := parseConfig([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Server.OAuth["oidc"].DisablePKCE != (value == "true") {
			t.Fatalf("environment override %q was not applied", value)
		}
	}
}
