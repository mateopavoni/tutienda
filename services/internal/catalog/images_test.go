package catalog

import "testing"

// TestNormalizeImages pins the gallery write-time invariants: the gallery is the unique, non-blank list
// of [imageUrl, ...images]; the primary (ImageURL) is always Images[0]; and the count is capped.
func TestNormalizeImages(t *testing.T) {
	t.Run("dedupes, trims and pins primary first", func(t *testing.T) {
		p := Product{ImageURL: "  a  ", Images: []string{"a", "", "b", "b", "c"}}
		normalizeImages(&p)
		if got, want := p.Images, []string{"a", "b", "c"}; !equal(got, want) {
			t.Fatalf("gallery = %v, want %v", got, want)
		}
		if p.ImageURL != "a" {
			t.Errorf("primary = %q, want first gallery image %q", p.ImageURL, "a")
		}
	})

	t.Run("primary defaults to first uploaded image when ImageURL is empty", func(t *testing.T) {
		p := Product{Images: []string{"x", "y"}}
		normalizeImages(&p)
		if p.ImageURL != "x" {
			t.Errorf("primary = %q, want %q", p.ImageURL, "x")
		}
	})

	t.Run("legacy single ImageURL becomes a one-item gallery", func(t *testing.T) {
		p := Product{ImageURL: "only"}
		normalizeImages(&p)
		if len(p.Images) != 1 || p.Images[0] != "only" {
			t.Errorf("gallery = %v, want [only]", p.Images)
		}
	})

	t.Run("caps the gallery", func(t *testing.T) {
		imgs := make([]string, maxProductImages+5)
		for i := range imgs {
			imgs[i] = string(rune('a' + i))
		}
		p := Product{Images: imgs}
		normalizeImages(&p)
		if len(p.Images) != maxProductImages {
			t.Errorf("gallery len = %d, want %d", len(p.Images), maxProductImages)
		}
	})
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
