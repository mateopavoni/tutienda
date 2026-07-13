package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrReservationNotFound is returned when a reservation id does not exist.
var ErrReservationNotFound = errors.New("reservation not found")

// Repository is the MongoDB persistence layer for stock and reservations.
type Repository struct {
	stocks       *mongo.Collection
	reservations *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		stocks:       db.Collection("stocks"),
		reservations: db.Collection("reservations"),
	}
}

// EnsureIndexes creates the indexes the service relies on:
//   - a unique compound {tenantId, sku} on stocks so a SKU is one document per store (and the same SKU
//     can exist independently in different stores);
//   - a plain (non-TTL) {status, expiresAt} on reservations to make the janitor's expiry query
//     efficient. It is deliberately NOT a TTL index: a TTL index would *delete* an expired reservation
//     and silently leak the held stock. The janitor must compensate stock before removing the doc.
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.stocks.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenantId", Value: 1}, {Key: "sku", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	_, err := r.reservations.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}, {Key: "expiresAt", Value: 1}},
	})
	return err
}

// deleteByTenant removes every stock and reservation document of a store (used when the store itself is
// deleted). Reservations aren't compensated first: the store is gone, so there's nothing left to protect.
func (r *Repository) deleteByTenant(ctx context.Context, tenant string) error {
	if _, err := r.stocks.DeleteMany(ctx, bson.M{"tenantId": tenant}); err != nil {
		return err
	}
	_, err := r.reservations.DeleteMany(ctx, bson.M{"tenantId": tenant})
	return err
}

// TryReserve atomically moves qty from available to reserved for one store's SKU, but only if enough is
// available. The {tenantId, sku, available: {$gte: qty}} guard means the update matches zero documents
// when stock is short, so the operation fails cleanly instead of driving stock negative. This is the
// single-document atomic primitive the whole no-oversell guarantee rests on — now scoped per store, so
// one store's load can never touch another's stock. No app lock, no transaction.
func (r *Repository) TryReserve(ctx context.Context, tenant, sku string, qty int) (bool, error) {
	res, err := r.stocks.UpdateOne(ctx,
		bson.M{"tenantId": tenant, "sku": sku, "available": bson.M{"$gte": qty}},
		bson.M{"$inc": bson.M{"available": -qty, "reserved": qty}},
	)
	if err != nil {
		return false, fmt.Errorf("try reserve %s: %w", sku, err)
	}
	return res.MatchedCount == 1, nil
}

// release returns reserved stock back to available.
func (r *Repository) release(ctx context.Context, tenant, sku string, qty int) error {
	_, err := r.stocks.UpdateOne(ctx,
		bson.M{"tenantId": tenant, "sku": sku},
		bson.M{"$inc": bson.M{"available": qty, "reserved": -qty}},
	)
	return err
}

// consume finalizes a sale: reserved stock leaves inventory for good.
func (r *Repository) consume(ctx context.Context, tenant, sku string, qty int) error {
	_, err := r.stocks.UpdateOne(ctx,
		bson.M{"tenantId": tenant, "sku": sku},
		bson.M{"$inc": bson.M{"reserved": -qty}},
	)
	return err
}

func (r *Repository) insertReservation(ctx context.Context, res *Reservation) (primitive.ObjectID, error) {
	out, err := r.reservations.InsertOne(ctx, res)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return out.InsertedID.(primitive.ObjectID), nil
}

// claim atomically transitions a reservation from one status to another, returning the reservation
// as it was *before* the change. It is the concurrency gate that guarantees a reservation is only
// ever finalized once: whoever wins the status transition (a user confirm or the janitor) is the
// only one allowed to touch stock for that reservation. A non-empty tenant scopes the claim to one
// store (user-driven confirm/release); the janitor passes "" because it already holds the document.
func (r *Repository) claim(ctx context.Context, id primitive.ObjectID, tenant string, from, to ReservationStatus) (*Reservation, error) {
	filter := bson.M{"_id": id, "status": from}
	if tenant != "" {
		filter["tenantId"] = tenant
	}
	var res Reservation
	err := r.reservations.FindOneAndUpdate(ctx,
		filter,
		bson.M{"$set": bson.M{"status": to}},
		options.FindOneAndUpdate().SetReturnDocument(options.Before),
	).Decode(&res)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil // someone else already claimed it, or it doesn't exist
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) getReservation(ctx context.Context, id primitive.ObjectID) (*Reservation, error) {
	var res Reservation
	err := r.reservations.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrReservationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// findExpiredHeld returns reservations still HELD whose hold has expired.
func (r *Repository) findExpiredHeld(ctx context.Context, now time.Time, limit int64) ([]Reservation, error) {
	cur, err := r.reservations.Find(ctx,
		bson.M{"status": StatusHeld, "expiresAt": bson.M{"$lt": now}},
		options.Find().SetLimit(limit),
	)
	if err != nil {
		return nil, err
	}
	var out []Reservation
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) deleteReservation(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.reservations.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *Repository) getStocks(ctx context.Context, tenant string, skus []string) ([]Stock, error) {
	cur, err := r.stocks.Find(ctx, bson.M{"tenantId": tenant, "sku": bson.M{"$in": skus}})
	if err != nil {
		return nil, err
	}
	var out []Stock
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// countStocks returns how many stock documents exist (used to seed on first boot only).
func (r *Repository) countStocks(ctx context.Context) (int64, error) {
	return r.stocks.CountDocuments(ctx, bson.M{})
}

// upsertStock sets the on-hand amount for a store's SKU (used by seeding and the admin stock editor).
// The {tenantId, sku} filter + $setOnInsert keeps it idempotent: re-running resets availability but
// never clobbers an existing doc's reserved count.
func (r *Repository) upsertStock(ctx context.Context, tenant, sku string, available int) error {
	_, err := r.stocks.UpdateOne(ctx,
		bson.M{"tenantId": tenant, "sku": sku},
		// tenantId and sku come from the filter's equality on insert; only set the rest here to avoid
		// a Mongo "conflict at 'tenantId'" error.
		bson.M{
			"$set":         bson.M{"available": available},
			"$setOnInsert": bson.M{"reserved": 0},
		},
		options.Update().SetUpsert(true),
	)
	return err
}
