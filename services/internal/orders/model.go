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

// Order is the persisted result of a checkout, owned by one store.
type Order struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID      string             `bson:"tenantId" json:"-"`
	Items         []Item             `bson:"items" json:"items"`
	Currency      string             `bson:"currency" json:"currency"`
	TotalCents    int64              `bson:"totalCents" json:"totalCents"`
	Status        OrderStatus        `bson:"status" json:"status"`
	ReservationID string             `bson:"reservationId" json:"reservationId"`
	FailureReason string             `bson:"failureReason,omitempty" json:"failureReason,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
}
