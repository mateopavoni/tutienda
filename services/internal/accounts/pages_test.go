package accounts

import (
	"strings"
	"testing"

	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
)

// TestSanitizePages pins the write-time invariants of standalone storefront pages: unknown types are
// dropped, duplicates collapse to the first occurrence, title/body are trimmed and capped, and a page
// with an empty body is dropped (an empty page is treated as absent — not routed or linked).
func TestSanitizePages(t *testing.T) {
	in := []Page{
		{Type: "about", Title: "  Our story  ", Body: "  We started in a garage.  "},
		{Type: "bogus", Title: "ignored", Body: "ignored"}, // unknown type → dropped
		{Type: "about", Title: "dup", Body: "dup"},          // duplicate → dropped
		{Type: "faq", Title: "FAQ", Body: ""},               // empty body → dropped
		{Type: "terms", Title: "Terms", Body: strings.Repeat("x", pageBodyMax+50)},
	}

	out := sanitizePages(in, plan.Scale)

	if len(out) != 2 {
		t.Fatalf("want 2 pages (about, terms), got %d: %+v", len(out), out)
	}
	if out[0].Type != "about" || out[0].Title != "Our story" || out[0].Body != "We started in a garage." {
		t.Errorf("about not trimmed/kept: %+v", out[0])
	}
	if out[1].Type != "terms" || len([]rune(out[1].Body)) != pageBodyMax {
		t.Errorf("terms body not capped to %d runes: %d", pageBodyMax, len([]rune(out[1].Body)))
	}
}

func TestSanitizePagesEmpty(t *testing.T) {
	if got := sanitizePages(nil, plan.Scale); got != nil {
		t.Errorf("nil in → nil out, got %v", got)
	}
	if got := sanitizePages([]Page{{Type: "about", Body: "  "}}, plan.Scale); got != nil {
		t.Errorf("all-empty-body in → nil out, got %v", got)
	}
}

// TestSanitizePagesTierGating pins that standalone pages require plan.FeatureStaticPages: free tier
// drops all pages regardless of content, pro/scale keep them.
func TestSanitizePagesTierGating(t *testing.T) {
	in := []Page{{Type: "about", Title: "About", Body: "We started in a garage."}}

	if got := sanitizePages(in, plan.Free); got != nil {
		t.Errorf("free tier must drop all standalone pages, got %v", got)
	}
	if got := sanitizePages(in, plan.Pro); len(got) != 1 {
		t.Errorf("pro tier must keep standalone pages, got %v", got)
	}
}
