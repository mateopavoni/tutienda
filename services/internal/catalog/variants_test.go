package catalog

import "testing"

// TestNormalizeVariants pins the write-time invariants of a multi-variant product: fields are trimmed,
// blank-SKU rows are dropped, duplicate SKUs are collapsed (they'd otherwise fight over one stock doc),
// and an empty label falls back to the SKU.
func TestNormalizeVariants(t *testing.T) {
	t.Run("trims, drops blank SKUs and dedupes", func(t *testing.T) {
		p := Product{Variants: []Variant{
			{SKU: "  A-S  ", Label: "  Small  "},
			{SKU: "", Label: "ignored"},      // no SKU → dropped
			{SKU: "A-M", Label: ""},          // blank label → falls back to SKU
			{SKU: "A-S", Label: "dup small"}, // duplicate SKU → dropped
		}}
		normalizeVariants(&p)
		if len(p.Variants) != 2 {
			t.Fatalf("variants = %d, want 2 (%v)", len(p.Variants), p.Variants)
		}
		if p.Variants[0].SKU != "A-S" || p.Variants[0].Label != "Small" {
			t.Errorf("variant[0] = %+v, want {A-S Small}", p.Variants[0])
		}
		if p.Variants[1].SKU != "A-M" || p.Variants[1].Label != "A-M" {
			t.Errorf("variant[1] = %+v, want label defaulted to SKU", p.Variants[1])
		}
	})

	t.Run("no variants stays empty (product can be authored before stock)", func(t *testing.T) {
		p := Product{}
		normalizeVariants(&p)
		if len(p.Variants) != 0 {
			t.Errorf("variants = %v, want empty", p.Variants)
		}
	})
}
