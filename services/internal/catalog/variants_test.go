package catalog

import (
	"strings"
	"testing"
)

// TestNormalizeVariants pins the write-time invariants of a multi-variant product: fields are trimmed, a
// blank SKU is auto-generated from the product name + variant label (still editable by the merchant
// before saving), duplicate SKUs are collapsed (they'd otherwise fight over one stock doc), and an empty
// label falls back to the SKU.
func TestNormalizeVariants(t *testing.T) {
	t.Run("trims, dedupes and leaves explicit SKUs alone", func(t *testing.T) {
		p := Product{Name: "Aero Jacket", Variants: []Variant{
			{SKU: "  A-S  ", Label: "  Small  "},
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

	t.Run("blank SKU is auto-generated from name + label", func(t *testing.T) {
		p := Product{Name: "Concrete Lamp 01", Variants: []Variant{
			{SKU: "", Label: "Black"},
			{SKU: "", Label: ""}, // no label either → falls back to the product name alone
		}}
		normalizeVariants(&p)
		if len(p.Variants) != 2 {
			t.Fatalf("variants = %d, want 2 (%v)", len(p.Variants), p.Variants)
		}
		if p.Variants[0].SKU != "CONCRETE-LAMP-01-BLACK" {
			t.Errorf("variant[0].SKU = %q, want CONCRETE-LAMP-01-BLACK", p.Variants[0].SKU)
		}
		if p.Variants[1].SKU != "CONCRETE-LAMP-01" {
			t.Errorf("variant[1].SKU = %q, want CONCRETE-LAMP-01", p.Variants[1].SKU)
		}
	})

	t.Run("auto-generated SKUs that collide still dedupe", func(t *testing.T) {
		p := Product{Name: "Tote", Variants: []Variant{
			{SKU: "", Label: ""},
			{SKU: "", Label: ""},
		}}
		normalizeVariants(&p)
		if len(p.Variants) != 2 {
			t.Fatalf("variants = %d, want 2 (%v)", len(p.Variants), p.Variants)
		}
		if p.Variants[0].SKU != "TOTE" || p.Variants[1].SKU != "TOTE-2" {
			t.Errorf("SKUs = %q, %q, want TOTE, TOTE-2", p.Variants[0].SKU, p.Variants[1].SKU)
		}
	})

	t.Run("no variants stays empty (product can be authored before stock)", func(t *testing.T) {
		p := Product{}
		normalizeVariants(&p)
		if len(p.Variants) != 0 {
			t.Errorf("variants = %v, want empty", p.Variants)
		}
	})

	t.Run("VariantLabel is trimmed and capped", func(t *testing.T) {
		p := Product{VariantLabel: "  Almacenamiento  "}
		normalizeVariants(&p)
		if p.VariantLabel != "Almacenamiento" {
			t.Errorf("VariantLabel = %q, want trimmed \"Almacenamiento\"", p.VariantLabel)
		}

		long := Product{VariantLabel: strings.Repeat("x", variantLabelMax+10)}
		normalizeVariants(&long)
		if len(long.VariantLabel) != variantLabelMax {
			t.Errorf("VariantLabel len = %d, want capped at %d", len(long.VariantLabel), variantLabelMax)
		}
	})
}
