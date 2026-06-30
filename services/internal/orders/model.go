package orders

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderStatus tracks the saga outcome.
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"   // reserved, awaiting payment confirmation
	StatusConfirmed OrderStatus = "CONFIRMED" // payment confirmed, stock consumed
	StatusFailed    OrderStatus = "FAILED"    // confirmation failed, stock released
)

// Item is a priced line of an order. Name and PriceCents are resolved server-side from the catalog,
// never taken from the client.
type Item struct {
	SKU        string `bson:"sku" json:"sku"`
	Qty        int    `bson:"qty" json:"qty"`
	Name       string `bson:"name" json:"name"`
	PriceCents int64  `bson:"priceCents" json:"priceCents"`
}

// ShippingAddress is where the order ships, collected at checkout. Merchant-facing only.
type ShippingAddress struct {
	Name       string `bson:"name" json:"name"`
	Line1      string `bson:"line1" json:"line1"`
	City       string `bson:"city" json:"city"`
	PostalCode string `bson:"postalCode" json:"postalCode"`
	Country    string `bson:"country" json:"country"`
}

// PaymentStatus is the (simulated) gateway outcome. This is a portfolio project — no real gateway is
// integrated; the storefront's fake card form drives whether the saga confirms or releases the hold.
type PaymentStatus string

const (
	PaymentPaid   PaymentStatus = "paid"
	PaymentFailed PaymentStatus = "failed"
)

// Order is the persisted result of a checkout, owned by one store.
type Order struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID      string             `bson:"tenantId" json:"-"`
	// CustomerID attributes the order to a logged-in storefront customer. Empty = a guest checkout
	// (accounts are optional). Not exposed in JSON; it's an internal ownership key for order history.
	CustomerID    string             `bson:"customerId,omitempty" json:"-"`
	Items         []Item             `bson:"items" json:"items"`
	Shipping      ShippingAddress    `bson:"shipping" json:"shipping"`
	Currency      string             `bson:"currency" json:"currency"`
	// TotalCents = items subtotal + ShippingCents (server-computed, never trusted from the client).
	ShippingCents int64              `bson:"shippingCents" json:"shippingCents"`
	TotalCents    int64              `bson:"totalCents" json:"totalCents"`
	Status        OrderStatus        `bson:"status" json:"status"`
	PaymentStatus PaymentStatus      `bson:"paymentStatus,omitempty" json:"paymentStatus,omitempty"`
	ReservationID string             `bson:"reservationId" json:"reservationId"`
	FailureReason string             `bson:"failureReason,omitempty" json:"failureReason,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}
