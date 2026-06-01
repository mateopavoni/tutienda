package accounts

import (
	"strings"
	"testing"
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

	out := sanitizeLayout(in)

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
		out := sanitizeLayout([]Section{{Type: typ, Enabled: true, Heading: "h", Body: "b"}})
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
	out := sanitizeLayout(in)
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
	capped := sanitizeLayout([]Section{{Type: "gallery", Enabled: true, Items: many}})
	if got := len(capped[0].Items); got != maxSectionItems {
		t.Errorf("items not capped: got %d, want %d", got, maxSectionItems)
	}
}

// TestSanitizeLayoutEmpty keeps an empty/absent layout as nil so the storefront falls back to its default
// (just the grid) instead of persisting an empty array.
func TestSanitizeLayoutEmpty(t *testing.T) {
	if out := sanitizeLayout(nil); out != nil {
		t.Errorf("nil layout should stay nil, got %+v", out)
	}
	if out := sanitizeLayout([]Section{{Type: "bogus"}}); len(out) != 0 {
		t.Errorf("all-unknown layout should sanitize to empty, got %+v", out)
	}
}
