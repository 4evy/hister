// SPDX-License-Identifier: AGPL-3.0-or-later

package languages

import (
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  Resolution
	}{
		{" EN-us ", Resolution{"en-US", "en", "en", "en", "en"}},
		{"no-NO", Resolution{"nb-NO", "nb", "nb", "no", "no"}},
		{"in-ID", Resolution{"id-ID", "id", "id", "default", ""}},
		{"zh-hant-TW", Resolution{"zh-Hant-TW", "zh", "zh", "cjk", "zh"}},
		{"uk-UA", Resolution{"uk-UA", "uk", "", "default", ""}},
		{"ckb", Resolution{"ckb", "ckb", "", "ckb", "ckb"}},
		{"cs", Resolution{"cs", "cs", "", "default", ""}},
		{"UNKNOWN", Resolution{"unknown", "unknown", "", "default", ""}},
		{"und", Resolution{"unknown", "unknown", "", "default", ""}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := Resolve(tc.input)
			if err != nil || got != tc.want {
				t.Fatalf("Resolve = %+v, %v; want %+v", got, err, tc.want)
			}
			if got.Index != "" && !IsIndexCode(got.Index) {
				t.Fatalf("unsafe index code %q", got.Index)
			}
		})
	}
}

func TestRejectInvalidLanguageIdentity(t *testing.T) {
	for _, value := range []string{"", " ", "zz", "../en", "en/US", "en_US", "en--US", "en-zzzzzzz", "auto", "default", "standard", "x-custom", "qaa", "und-Latn", "en-x-custom", "en-u-ca-gregory", strings.Repeat("a", 129)} {
		if _, err := Resolve(value); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	for _, code := range []string{"nb", "en-US", "uk", "../en", "zz", "standard", "EN"} {
		if IsIndexCode(code) {
			t.Errorf("accepted index code %q", code)
		}
	}
}
