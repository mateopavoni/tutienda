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
	// Layout is the ordered list of storefront sections the merchant arranged. Empty = the default
	// layout (just the product grid). The grid ("catalog") is mandatory; the rest are optional blocks.
	Layout []Section `bson:"layout,omitempty" json:"layout,omitempty"`
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
	Plan      string    `bson:"plan" json:"plan"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
