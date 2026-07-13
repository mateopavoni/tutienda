// Package plan is the single source of truth for the SaaS membership tiers and what each one unlocks.
// A store carries a Tier; the store-scoped JWT signs it, the gateway passes it downstream as
// X-Store-Plan, and the services consult this registry to decide whether a paid feature (a "tool"/
// plugin) is available. No other package hardcodes "pro can do X" — they ask Tier.Has / Tier.ProductLimit,
// so adding a feature or repricing a tier is a one-file change here.
package plan

import (
	"context"
	"net/http"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
)

// Header is the wire name the gateway uses to pass the store's plan downstream (mirrors tenant.Header).
const Header = "X-Store-Plan"

// Tier is a membership level. The zero value is unusable on purpose; use Normalize on untrusted input.
type Tier string

const (
	Free  Tier = "free"
	Pro   Tier = "pro"
	Scale Tier = "scale"
)

// Feature is a gateable capability — a paid tool/plugin a merchant unlocks by upgrading. The string is
// stable wire/UI identity, so the frontend can mirror these names when it dims a locked control.
type Feature string

const (
	// FeatureDrops gates scheduled drops (a future dropAt) with the live stock counter on the storefront.
	FeatureDrops Feature = "drops"
	// FeatureCustomLogo gates uploading a store logo. Every tier may set an accent (with the contrast guard).
	FeatureCustomLogo Feature = "customLogo"
	// FeatureAnalytics gates the (future) sales analytics dashboard.
	FeatureAnalytics Feature = "analytics"
	// FeatureRemoveBranding gates hiding the "powered by" platform mark from the storefront footer.
	FeatureRemoveBranding Feature = "removeBranding"
	// FeatureSectionsPro gates the gallery/faq/feature/contact storefront section types (free keeps
	// announcement/hero/catalog/about only).
	FeatureSectionsPro Feature = "sectionsPro"
	// FeatureStaticPages gates the standalone about/faq/contact/terms pages.
	FeatureStaticPages Feature = "staticPages"
)

// Unlimited is the sentinel ProductLimit for a tier with no product cap.
const Unlimited = -1

type entitlement struct {
	features     map[Feature]bool
	productLimit int
}

// registry is the entitlement table. Each tier lists exactly the features it unlocks; anything not
// listed is locked. Keep this the only place that maps tiers to capabilities.
var registry = map[Tier]entitlement{
	Free: {
		features:     map[Feature]bool{},
		productLimit: 15,
	},
	Pro: {
		features: map[Feature]bool{
			FeatureDrops:       true,
			FeatureCustomLogo:  true,
			FeatureAnalytics:   true,
			FeatureSectionsPro: true,
			FeatureStaticPages: true,
		},
		productLimit: 300,
	},
	Scale: {
		features: map[Feature]bool{
			FeatureDrops:          true,
			FeatureCustomLogo:     true,
			FeatureAnalytics:      true,
			FeatureRemoveBranding: true,
			FeatureSectionsPro:    true,
			FeatureStaticPages:    true,
		},
		productLimit: Unlimited,
	},
}

// All returns the tiers in upgrade order (cheapest first). Used by the API to advertise plans.
func All() []Tier { return []Tier{Free, Pro, Scale} }

// Normalize coerces arbitrary/empty input to a known tier, defaulting to Free. An unknown string maps
// to Free rather than erroring, so a typo or a tampered claim can never silently unlock a paid feature.
func Normalize(s string) Tier {
	t := Tier(s)
	if _, ok := registry[t]; ok {
		return t
	}
	return Free
}

// Valid reports whether s is exactly one of the known tiers (used to reject bad upgrade requests).
func Valid(s string) bool {
	_, ok := registry[Tier(s)]
	return ok
}

// Has reports whether the tier unlocks the feature. It normalizes first, so an unknown tier has nothing.
func (t Tier) Has(f Feature) bool {
	return registry[Normalize(string(t))].features[f]
}

// ProductLimit returns the max products a store on this tier may hold, or Unlimited (-1).
func (t Tier) ProductLimit() int {
	return registry[Normalize(string(t))].productLimit
}

// Features returns the unlocked features of the tier (unordered), handy for the UI/entitlements API.
func (t Tier) Features() []Feature {
	ent := registry[Normalize(string(t))]
	out := make([]Feature, 0, len(ent.features))
	for f := range ent.features {
		out = append(out, f)
	}
	return out
}

type ctxKey struct{}

// WithPlan returns a child context carrying the active store's plan.
func WithPlan(ctx context.Context, t Tier) context.Context {
	return context.WithValue(ctx, ctxKey{}, t)
}

// FromContext returns the active plan, defaulting to Free when none was stamped (a store-scoped admin
// request always carries one; defaulting to Free keeps an unauthenticated path from unlocking anything).
func FromContext(ctx context.Context) Tier {
	t, ok := ctx.Value(ctxKey{}).(Tier)
	if !ok {
		return Free
	}
	return t
}

// Middleware lifts the X-Store-Plan header into the request context, normalizing it. Like the tenant
// middleware it never rejects — a missing plan simply reads back as Free downstream.
func Middleware() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if p := r.Header.Get(Header); p != "" {
				r = r.WithContext(WithPlan(r.Context(), Normalize(p)))
			}
			next.ServeHTTP(w, r)
		})
	}
}
