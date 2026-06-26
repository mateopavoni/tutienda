package accounts

import (
	"strings"
	"testing"
)

// TestSlugRe pins the store-slug rule (3-40 chars, lowercase alphanumeric/hyphen, no leading/trailing
// hyphen) that CreateStore enforces before touching the DB. A slug becomes a public subdomain, so the
// shape matters and is worth a DB-free unit.
func TestSlugRe(t *testing.T) {
	valid := []string{"abc", "a-b", "system-archive", "store123", strings.Repeat("a", 40)}
	for _, s := range valid {
		if !slugRe.MatchString(s) {
			t.Errorf("want valid: %q", s)
		}
	}
	// empty, too short (<3), leading/trailing hyphen, uppercase, space, non-ascii, too long (>40).
	invalid := []string{"", "ab", "-abc", "abc-", "AbC", "a b", "acmé", strings.Repeat("a", 41)}
	for _, s := range invalid {
		if slugRe.MatchString(s) {
			t.Errorf("want invalid: %q", s)
		}
	}
}

// TestCapRunes makes sure truncation counts runes, not bytes, so a multi-byte character is never split
// in half (which would corrupt the stored copy).
func TestCapRunes(t *testing.T) {
	if got := capRunes("hello", 10); got != "hello" {
		t.Errorf("under cap unchanged: got %q", got)
	}
	if got := capRunes("hello", 3); got != "hel" {
		t.Errorf("ascii cap: got %q", got)
	}
	// 5 multi-byte runes capped to 3 → 3 whole runes (6 bytes), not 3 bytes (which would split a rune).
	if got := capRunes("ñññññ", 3); got != "ñññ" {
		t.Errorf("multibyte cap: got %d runes %q", len([]rune(got)), got)
	}
}
