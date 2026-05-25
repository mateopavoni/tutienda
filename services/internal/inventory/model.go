package inventory

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Stock is the amount of a SKU on hand for one store. Invariant: Available >= 0 at all times.
//   - Available: free to reserve.
//   - Reserved: held by an open reservation (not yet sold, not free).
//
// Identity is the compound (TenantID, SKU): a SKU string is only unique within a store, so the same SKU
// can exist independently in two stores. A unique index on {tenantId, sku} enforces one doc per pair;
// the Mongo _id is an opaque ObjectID.
type Stock struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	TenantID  string             `bson:"tenantId" json:"-"`
	SKU       string             `bson:"sku" json:"sku"`
	Available int                `bson:"available" json:"available"`
	Reserved  int                `bson:"reserved" json:"reserved"`
}

// Line is one SKU + quantity within a reservation.
type Line struct {
	SKU string `bson:"sku" json:"sku"`
	Qty int    `bson:"qty" json:"qty"`
}

// ReservationStatus is the lifecycle of a hold on stock.
type ReservationStatus string

const (
	StatusHeld      ReservationStatus = "HELD"      // stock is held, awaiting confirm or expiry
	StatusConfirmed ReservationStatus = "CONFIRMED" // goods sold; reserved decremented for good
	StatusReleased  ReservationStatus = "RELEASED"  // stock returned to available
)

// Reservation is a short-lived hold over one or more SKUs, owned by a single store.
type Reservation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TenantID  string             `bson:"tenantId" json:"-"`
	Lines     []Line             `bson:"lines" json:"lines"`
	Status    ReservationStatus  `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
}
