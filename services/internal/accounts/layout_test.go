package accounts

import (
	"strings"
	"testing"

	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
)

// TestSanitizeLayout pins the write-time invariants of a storefront layout: unknown/legacy section types
// are dropped, duplicates collapse to the first occurrence, the mandatory product grid is forced enabled
// even if a client tried to hide it, and oversized copy is capped. The storefront also normalizes at
// render time, but sanitizing on write is what keeps the stored document trustworthy.
func TestSanitizeLayout(t *testing.T) {
	in := []Section{
		{Type: "hero", Enabled: true, Heading: "  New season  ", Body: "live"},
		{Type: "bogus", Enabled: true, Heading: "ignored"}, // unknown type → dropped
		{Type: "hero", Enabled: false},                     // duplicate → dropped
		{Type: "catalog", Enabled: false},                  // grid → forced back on
		{Type: "about", Enabled: true, Body: strings.Repeat("x", sectionFieldMax+50)},
	}

	out := sanitizeLayout(in, plan.Scale)

	if len(out) != 3 {
		t.Fatalf("want 3 sections (hero, catalog, about), got %d: %+v", len(out), out)
	}
	if out[0].Type != "hero" || out[1].Type != "catalog" || out[2].Type != "about" {
		t.Fatalf("order not preserved: %+v", out)
	}
	if out[0].Heading != "New season" {
		t.Errorf("heading not trimmed: %q", out[0].Heading)
	}
	if !out[1].Enabled {
		t.Error("catalog must be forced enabled (mandatory grid)")
	}
	if got := len([]rune(out[2].Body)); got != sectionFieldMax {
		t.Errorf("body not capped: got %d runes, want %d", got, sectionFieldMax)
	}
}

// TestSanitizeLayoutKnownTypes pins that every supported block type survives sanitization, so adding a
// section type to knownSectionTypes (and to the TS mirror) is enough to make it usable end to end.
func TestSanitizeLayoutKnownTypes(t *testing.T) {
	for _, typ := range []string{"announcement", "hero", "feature", "catalog", "about", "contact", "gallery", "faq"} {
		out := sanitizeLayout([]Section{{Type: typ, Enabled: true, Heading: "h", Body: "b"}}, plan.Scale)
		if len(out) != 1 || out[0].Type != typ {
			t.Errorf("known type %q should survive sanitize, got %+v", typ, out)
		}
	}
}

// TestSanitizeItems pins the repeatable-item invariants: fully-empty items are dropped, copy is trimmed,
// and the per-section item count is capped so a section can't smuggle an unbounded list into the document.
func TestSanitizeItems(t *testing.T) {
	in := []Section{{
		Type:    "faq",
		Enabled: true,
		Items: []SectionItem{
			{Heading: "  Shipping?  ", Body: "2 days"}, // trimmed, kept
			{},                                         // fully empty → dropped
			{Body: "answer only"},                      // partial → kept
		},
	}}
	out := sanitizeLayout(in, plan.Scale)
	if len(out) != 1 || len(out[0].Items) != 2 {
		t.Fatalf("want 1 section with 2 items (empty dropped), got %+v", out)
	}
	if out[0].Items[0].Heading != "Shipping?" {
		t.Errorf("item heading not trimmed: %q", out[0].Items[0].Heading)
	}

	// More items than the cap collapse to maxSectionItems.
	many := make([]SectionItem, maxSectionItems+10)
	for i := range many {
		many[i] = SectionItem{Body: "x"}
	}
	capped := sanitizeLayout([]Section{{Type: "gallery", Enabled: true, Items: many}}, plan.Scale)
	if got := len(capped[0].Items); got != maxSectionItems {
		t.Errorf("items not capped: got %d, want %d", got, maxSectionItems)
	}
}

// TestSanitizeLayoutHeroOverlay pins the hero overlay invariants: an unset/out-of-range strength clamps
// into [heroOverlayMin, heroOverlayMax] (missing defaults to heroOverlayDefault), and an unknown/tampered
// color falls back to "black" — same defensive stance as an unknown plan tier or section type.
func TestSanitizeLayoutHeroOverlay(t *testing.T) {
	cases := []struct {
		name     string
		in       Section
		wantPct  int
		wantColo string
	}{
		{"unset defaults", Section{Type: "hero"}, heroOverlayDefault, "black"},
		{"below min clamps up", Section{Type: "hero", Overlay: 1}, heroOverlayMin, "black"},
		{"above max clamps down", Section{Type: "hero", Overlay: 999}, heroOverlayMax, "black"},
		{"in range kept", Section{Type: "hero", Overlay: 55, OverlayColor: "white"}, 55, "white"},
		{"bogus color falls back", Section{Type: "hero", OverlayColor: "chartreuse"}, heroOverlayDefault, "black"},
		{"black kept", Section{Type: "hero", Overlay: 60, OverlayColor: "black"}, 60, "black"},
	}
	for _, c := range cases {
		out := sanitizeLayout([]Section{c.in}, plan.Scale)
		if len(out) != 1 {
			t.Fatalf("%s: want 1 section, got %+v", c.name, out)
		}
		if out[0].Overlay != c.wantPct {
			t.Errorf("%s: overlay = %d, want %d", c.name, out[0].Overlay, c.wantPct)
		}
		if out[0].OverlayColor != c.wantColo {
			t.Errorf("%s: overlayColor = %q, want %q", c.name, out[0].OverlayColor, c.wantColo)
		}
	}
}

// TestSanitizeLayoutEmpty keeps an empty/absent layout as nil so the storefront falls back to its default
// (just the grid) instead of persisting an empty array.
func TestSanitizeLayoutEmpty(t *testing.T) {
	if out := sanitizeLayout(nil, plan.Scale); out != nil {
		t.Errorf("nil layout should stay nil, got %+v", out)
	}
	if out := sanitizeLayout([]Section{{Type: "bogus"}}, plan.Scale); len(out) != 0 {
		t.Errorf("all-unknown layout should sanitize to empty, got %+v", out)
	}
}

// TestSanitizeLayoutTierGating pins that gallery/faq/feature/contact require plan.FeatureSectionsPro:
// free drops them (same treatment as an unknown type) while pro/scale keep them.
func TestSanitizeLayoutTierGating(t *testing.T) {
	in := []Section{
		{Type: "hero", Enabled: true},
		{Type: "feature", Enabled: true},
		{Type: "catalog", Enabled: true},
		{Type: "contact", Enabled: true},
		{Type: "gallery", Enabled: true},
		{Type: "faq", Enabled: true},
	}

	free := sanitizeLayout(in, plan.Free)
	if len(free) != 2 {
		t.Fatalf("free tier: want 2 sections (hero, catalog), got %d: %+v", len(free), free)
	}
	for _, sec := range free {
		if proSectionTypes[sec.Type] {
			t.Errorf("free tier must not keep pro section %q", sec.Type)
		}
	}

	pro := sanitizeLayout(in, plan.Pro)
	if len(pro) != len(in) {
		t.Fatalf("pro tier: want all %d sections kept, got %d: %+v", len(in), len(pro), pro)
	}
}
