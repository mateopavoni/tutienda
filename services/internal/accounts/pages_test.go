package accounts

import (
	"strings"
	"testing"

	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
)

// TestSanitizePages pins the write-time invariants of standalone storefront pages (just Terms now — see
// knownPageTypes): unknown types are dropped, duplicates collapse to the first occurrence, title/body are
// trimmed and capped, and a page with an empty body is dropped (an empty page is treated as absent — not
// routed or linked).
func TestSanitizePages(t *testing.T) {
	in := []Page{
		{Type: "terms", Title: "  Terms  ", Body: "  " + strings.Repeat("x", pageBodyMax+50) + "  "},
		{Type: "bogus", Title: "ignored", Body: "ignored"}, // unknown type → dropped
		{Type: "about", Title: "ignored", Body: "ignored"}, // no longer a known type → dropped
		{Type: "terms", Title: "dup", Body: "dup"},         // duplicate → dropped
	}

	out := sanitizePages(in, plan.Scale)

	if len(out) != 1 {
		t.Fatalf("want 1 page (terms), got %d: %+v", len(out), out)
	}
	if out[0].Type != "terms" || out[0].Title != "Terms" {
		t.Errorf("terms not trimmed/kept: %+v", out[0])
	}
	if len([]rune(out[0].Body)) != pageBodyMax {
		t.Errorf("terms body not capped to %d runes: %d", pageBodyMax, len([]rune(out[0].Body)))
	}
}

func TestSanitizePagesEmpty(t *testing.T) {
	if got := sanitizePages(nil, plan.Scale); got != nil {
		t.Errorf("nil in → nil out, got %v", got)
	}
	if got := sanitizePages([]Page{{Type: "terms", Body: "  "}}, plan.Scale); got != nil {
		t.Errorf("all-empty-body in → nil out, got %v", got)
	}
}

// TestSanitizePagesTierGating pins that standalone pages require plan.FeatureStaticPages: free tier
// drops all pages regardless of content, pro/scale keep them.
func TestSanitizePagesTierGating(t *testing.T) {
	in := []Page{{Type: "terms", Title: "Terms", Body: "Terms and conditions."}}

	if got := sanitizePages(in, plan.Free); got != nil {
		t.Errorf("free tier must drop all standalone pages, got %v", got)
	}
	if got := sanitizePages(in, plan.Pro); len(got) != 1 {
		t.Errorf("pro tier must keep standalone pages, got %v", got)
	}
}

// TestSanitizePagesSlug pins the route-slug derivation: an explicit slug is kept as-is (once valid), a
// blank one falls back to the title, then to the type, and two pages that would otherwise collide on the
// same slug never end up sharing a URL (duplicate Terms rows survive here only because sanitizePages'
// dedupe-by-type runs after slug assignment in this direct unit test of the slug step alone).
func TestSanitizePagesSlug(t *testing.T) {
	seen := map[string]bool{}
	got := pageSlug("quienes-somos", "Terms", "terms", seen)
	seen[got] = true
	if got != "quienes-somos" {
		t.Errorf("explicit slug not kept: %q", got)
	}

	got2 := pageSlug("", "My Terms", "terms", seen)
	seen[got2] = true
	if got2 != "my-terms" {
		t.Errorf("blank slug should fall back to slugified title: %q", got2)
	}

	got3 := pageSlug("", "My Terms", "terms", seen) // same title again → collides, bumps a suffix off the type
	if got3 != "terms" {
		t.Errorf("colliding title-slug should fall back to the type: %q", got3)
	}
}
