package accounts

import "testing"

// TestNormalizeTheme pins the defensive default: any unknown, empty, or tampered theme falls back to
// "monolith" (never silently renders an undefined template), same stance as an unknown plan tier → free.
// The TS mirror is web/src/lib/theme.test.ts.
func TestNormalizeTheme(t *testing.T) {
	cases := map[string]string{
		"":         defaultTheme,
		"  ":       defaultTheme,
		"hacker":   defaultTheme,
		"monolith": "monolith",
		"boutique": "boutique",
		"pop":      "pop",
		"terminal": "terminal",
		"  pop  ":  "pop",
	}
	for in, want := range cases {
		if got := normalizeTheme(in); got != want {
			t.Errorf("normalizeTheme(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestSanitizeHexAccent pins that only #rrggbb survives (lowercased); anything else ⇒ "" so the storefront
// falls back to the theme's own accent instead of persisting a junk/forced color.
func TestSanitizeHexAccent(t *testing.T) {
	cases := map[string]string{
		"#1B03EA":   "#1b03ea",
		"#6d40e1":   "#6d40e1",
		"  #abc123 ": "#abc123",
		"":          "",
		"red":       "",
		"#abc":      "", // shorthand not accepted
		"1b03ea":    "", // missing #
		"#1b03ea99": "", // too long
	}
	for in, want := range cases {
		if got := sanitizeHexAccent(in); got != want {
			t.Errorf("sanitizeHexAccent(%q) = %q, want %q", in, got, want)
		}
	}
}
