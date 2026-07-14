package catalog

import "time"

// Variant is a sellable option of a product, keyed by SKU (the unit inventory tracks stock against).
type Variant struct {
	SKU   string `bson:"sku" json:"sku"`
	Label string `bson:"label" json:"label"`
	Color string `bson:"color,omitempty" json:"color,omitempty"`
}

// Product is a catalog entry owned by one store. Prices are integer cents to avoid floating-point
// money bugs. Queries are always scoped by TenantID, so the same product id under another store is a
// different, isolated product.
type Product struct {
	ID          string `bson:"_id" json:"id"`
	TenantID    string `bson:"tenantId" json:"-"`
	Name        string `bson:"name" json:"name"`
	Category    string `bson:"category" json:"category"`
	Description string `bson:"description" json:"description"`
	PriceCents  int64  `bson:"priceCents" json:"priceCents"`
	Currency    string `bson:"currency" json:"currency"`
	// ImageURL is the primary image (used in the grid/card and as the cart thumbnail). Images is the full
	// gallery shown on the product detail page; ImageURL is kept as Images[0] for back-compat with the
	// card and the seed, which only know a single URL.
	ImageURL string    `bson:"imageUrl" json:"imageUrl"`
	Images   []string  `bson:"images,omitempty" json:"images,omitempty"`
	Variants []Variant `bson:"variants" json:"variants"`
	// VariantLabel names what a variant's Label actually is for this product (e.g. "Size", "Color",
	// "Storage") — the storefront selector shows it instead of a hardcoded "Size" so a phone store can say
	// "Storage" instead. Blank falls back to a generic "Size" client-side (old products, or a product with
	// no variants at all).
	VariantLabel string    `bson:"variantLabel,omitempty" json:"variantLabel,omitempty"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	// DropAt, when set and in the future, marks this product as a scheduled drop: the storefront shows a
	// countdown and the orders service refuses checkout until the release time. nil = on sale now (normal
	// product). This is the data half of "drop mode" — the live-stock counter on top of the no-oversell
	// engine is what makes a hyped, scarce release fair under a flash crowd.
	DropAt *time.Time `bson:"dropAt,omitempty" json:"dropAt,omitempty"`
}

// VariantInfo is the server-authoritative pricing answer the orders service asks for by SKU,
// so the client can never dictate the price it pays. DropAt rides along so the checkout saga can
// reject a purchase before a scheduled drop is released (gated server-side, not just in the UI).
type VariantInfo struct {
	SKU         string     `json:"sku"`
	ProductID   string     `json:"productId"`
	ProductName string     `json:"productName"`
	Label       string     `json:"label"`
	PriceCents  int64      `json:"priceCents"`
	Currency    string     `json:"currency"`
	DropAt      *time.Time `json:"dropAt,omitempty"`
}
