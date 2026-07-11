package accounts

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Merchant is a person who owns one or more stores. The password is only ever stored hashed. Role is
// the platform role (see authx.RoleMerchant/RoleAdmin); the SaaS super-admin is just a merchant with
// RoleAdmin, which is signed into their token and gates the cross-tenant /api/platform routes.
type Merchant struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Role         string             `bson:"role" json:"role"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}

// Settings is the per-store customization the storefront applies on top of the fixed design system.
// The visual language ("Monolith Editorial") does not change; only these knobs do.
type Settings struct {
	LogoURL     string `bson:"logoUrl,omitempty" json:"logoUrl,omitempty"`
	AccentColor string `bson:"accentColor,omitempty" json:"accentColor,omitempty"`
	Currency    string `bson:"currency,omitempty" json:"currency,omitempty"`
	// Theme is the storefront visual template (fonts + palette); see knownThemes and web/src/lib/theme.ts.
	// Unknown/empty ⇒ the default "monolith", same defensive stance as an unknown plan tier falling to free.
	Theme string `bson:"theme,omitempty" json:"theme,omitempty"`
	// Layout is the ordered list of storefront sections the merchant arranged. Empty = the default
	// layout (just the product grid). The grid ("catalog") is mandatory; the rest are optional blocks.
	Layout []Section `bson:"layout,omitempty" json:"layout,omitempty"`
	// Pages are standalone storefront pages, each at its own route (/store/{slug}/{type}). Unlike layout
	// sections (homepage blocks), these are full pages the merchant fills in: about/faq/contact/terms. A
	// page with an empty body is treated as absent — it isn't routed or linked.
	Pages []Page `bson:"pages,omitempty" json:"pages,omitempty"`
}

// Page is a standalone storefront page (about/faq/contact/terms). Type is the route segment; Title is the
// heading (falls back to a default per type in the UI); Body is free text rendered with line breaks kept.
type Page struct {
	Type  string `bson:"type" json:"type"`
	Title string `bson:"title,omitempty" json:"title,omitempty"`
	Body  string `bson:"body,omitempty" json:"body,omitempty"`
}

// Section is one configurable block of a store's storefront. A merchant chooses which optional sections
// appear, in what order, and their copy/image — but the typography, spacing and color tokens stay fixed.
// Type is one of: "announcement", "hero", "feature", "catalog" (mandatory grid), "about", "contact",
// "gallery", "faq". Unknown types are ignored by the storefront. A hero/feature CTA links to the catalog.
//
// Most sections use the scalar copy fields; the repeatable ones ("gallery", "faq") instead use Items —
// a gallery item is an image (ImageURL + optional Heading caption), an FAQ item is a Heading question
// plus a Body answer. Keeping both on one struct avoids a schema fork while the storefront/editor pick
// the relevant fields per type.
type Section struct {
	Type     string        `bson:"type" json:"type"`
	Enabled  bool          `bson:"enabled" json:"enabled"`
	Heading  string        `bson:"heading,omitempty" json:"heading,omitempty"`
	Body     string        `bson:"body,omitempty" json:"body,omitempty"`
	ImageURL string        `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
	CTALabel string        `bson:"ctaLabel,omitempty" json:"ctaLabel,omitempty"`
	Items    []SectionItem `bson:"items,omitempty" json:"items,omitempty"`
	// Overlay is the peak tint strength (0-100) of a gradient laid over a hero's background image, at the
	// edge nearest the text (the storefront's copy is left-aligned, so the gradient starts strong on the
	// left and fades to transparent towards the right — the photo stays visible, only the text's corner
	// is protected). OverlayColor picks the tint: "black" (default, white text) or "white" (dark text, for
	// a merchant who wants an airy/bright hero). Both merchant-adjustable; see sanitizeLayout for the
	// clamped range/closed set.
	Overlay      int    `bson:"overlay,omitempty" json:"overlay,omitempty"`
	OverlayColor string `bson:"overlayColor,omitempty" json:"overlayColor,omitempty"`
}

// SectionItem is one repeated entry inside an item-based section (a gallery image or an FAQ Q&A). The
// fields are reused by type: gallery → ImageURL (+ optional Heading caption); faq → Heading (question)
// and Body (answer).
type SectionItem struct {
	Heading  string `bson:"heading,omitempty" json:"heading,omitempty"`
	Body     string `bson:"body,omitempty" json:"body,omitempty"`
	ImageURL string `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
}

// Store is a tenant. Its id (ObjectID hex) is the tenantId threaded through every other service; the
// slug is the human-facing handle used to resolve the tenant from a subdomain/header.
type Store struct {
	ID          string    `bson:"_id" json:"id"`
	Slug        string    `bson:"slug" json:"slug"`
	OwnerID     string    `bson:"ownerId" json:"ownerId"`
	DisplayName string    `bson:"displayName" json:"displayName"`
	Settings    Settings  `bson:"settings" json:"settings"`
	// Plan is the store's membership tier (see platform/plan). It rides into the store-scoped JWT so the
	// gateway/services can gate paid features, and is exposed to the dashboard so it can dim locked tools.
	Plan string `bson:"plan" json:"plan"`
	// Disabled turns the storefront off without deleting anything (gateway returns 404 for its slug).
	// Named as the negative so the Go/Mongo zero value for every store created before this field existed
	// is "not disabled" — i.e. still active — with no migration needed.
	Disabled  bool      `bson:"disabled,omitempty" json:"disabled"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
