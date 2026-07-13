package plan

import "testing"

func TestNormalizeDefaultsToFree(t *testing.T) {
	cases := map[string]Tier{
		"free":    Free,
		"pro":     Pro,
		"scale":   Scale,
		"":        Free,
		"PRO":     Free, // case-sensitive on purpose: the claim is machine-written, not user input
		"unknown": Free,
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEntitlements(t *testing.T) {
	if Free.Has(FeatureDrops) {
		t.Error("free must not unlock drops")
	}
	if !Pro.Has(FeatureDrops) {
		t.Error("pro must unlock drops")
	}
	if !Scale.Has(FeatureRemoveBranding) {
		t.Error("scale must unlock removeBranding")
	}
	if Pro.Has(FeatureRemoveBranding) {
		t.Error("pro must not unlock removeBranding")
	}
	// An unknown/tampered tier must behave exactly like Free: nothing unlocked.
	if Tier("enterprise").Has(FeatureDrops) {
		t.Error("unknown tier must not unlock paid features")
	}
	if Free.Has(FeatureSectionsPro) {
		t.Error("free must not unlock sectionsPro")
	}
	if !Pro.Has(FeatureSectionsPro) {
		t.Error("pro must unlock sectionsPro")
	}
	if Free.Has(FeatureStaticPages) {
		t.Error("free must not unlock staticPages")
	}
	if !Scale.Has(FeatureStaticPages) {
		t.Error("scale must unlock staticPages")
	}
}

func TestProductLimits(t *testing.T) {
	if Free.ProductLimit() != 15 {
		t.Errorf("free product limit = %d, want 15", Free.ProductLimit())
	}
	if Scale.ProductLimit() != Unlimited {
		t.Errorf("scale product limit = %d, want Unlimited", Scale.ProductLimit())
	}
}
