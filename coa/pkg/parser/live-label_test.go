package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withBrandingDesc(t *testing.T, contents ...string) {
	t.Helper()
	original := brandingDescCandidates
	t.Cleanup(func() { brandingDescCandidates = original })

	dir := t.TempDir()
	paths := make([]string, 0, len(contents))
	for i, content := range contents {
		path := filepath.Join(dir, string(rune('a'+i)), "branding.desc")
		if content != "" {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
		paths = append(paths, path)
	}
	brandingDescCandidates = paths
}

func TestResolveLiveLabelFromWardrobeBranding(t *testing.T) {
	withBrandingDesc(t, "---\ncomponentName: eggs\nstrings:\n    productName:         Quirinux\n    shortProductName:    Quirinux\n")
	if got := ResolveLiveLabel(); got != "Quirinux" {
		t.Fatalf("ResolveLiveLabel() = %q, want Quirinux", got)
	}
}

func TestResolveLiveLabelFallsBackToShortNameAndSecondCandidate(t *testing.T) {
	withBrandingDesc(t,
		"strings:\n  productName: \"{{ distro }}\"\n  shortProductName: \"\"\n",
		"strings:\n  productName: \"QUIRINUX\"\n",
	)
	if got := ResolveLiveLabel(); got != "QUIRINUX" {
		t.Fatalf("ResolveLiveLabel() = %q, want QUIRINUX", got)
	}
}

func TestResolveLiveLabelDefaultsWithoutBranding(t *testing.T) {
	withBrandingDesc(t, "") // el fichero no existe
	if got := ResolveLiveLabel(); got != DefaultLiveLabel {
		t.Fatalf("ResolveLiveLabel() = %q, want %q", got, DefaultLiveLabel)
	}
}

func TestResolveLiveLabelIgnoresInvalidYAML(t *testing.T) {
	withBrandingDesc(t, "strings: [unterminated\n")
	if got := ResolveLiveLabel(); got != DefaultLiveLabel {
		t.Fatalf("ResolveLiveLabel() = %q, want %q", got, DefaultLiveLabel)
	}
}

func TestSanitizeLiveLabel(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Quirinux", "Quirinux"},
		{"Quirinux GNU/Linux", "Quirinux_GNU_Linux"},
		{"  My  Distro  ", "My_Distro"},
		{"Ñandú OS", "and_OS"},
		{"***", ""},
		{"", ""},
		{"OA_LIVE", "OA_LIVE"},
		{strings.Repeat("a", 50), strings.Repeat("a", 32)},
		{strings.Repeat("a", 31) + "/b", strings.Repeat("a", 31)},
	}
	for _, c := range cases {
		if got := SanitizeLiveLabel(c.in); got != c.want {
			t.Errorf("SanitizeLiveLabel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
