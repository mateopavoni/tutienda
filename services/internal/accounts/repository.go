package accounts

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// ErrMerchantNotFound is returned when an email has no matching merchant.
	ErrMerchantNotFound = errors.New("merchant not found")
	// ErrStoreNotFound is returned when a store id or slug has no match.
	ErrStoreNotFound = errors.New("store not found")
	// ErrDuplicate is returned when a unique constraint (email or slug) is violated.
	ErrDuplicate = errors.New("already exists")
)

// Repository persists merchants, stores, storefront customers and their contact-form messages.
type Repository struct {
	merchants *mongo.Collection
	stores    *mongo.Collection
	customers *mongo.Collection
	messages  *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		merchants: db.Collection("merchants"),
		stores:    db.Collection("stores"),
		customers: db.Collection("customers"),
		messages:  db.Collection("messages"),
	}
}

// EnsureIndexes enforces unique email per merchant, unique slug per store, and a unique email *per store*
// for customers — the same address can register independently in two different stores (a customer of
// store A is unrelated to one of store B), so the customer uniqueness key is the {tenantId, email} pair.
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	if _, err := r.merchants.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := r.stores.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := r.customers.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tenantId", Value: 1}, {Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	_, err := r.messages.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "tenantId", Value: 1}, {Key: "createdAt", Value: -1}},
	})
	return err
}

func (r *Repository) insertMerchant(ctx context.Context, m *Merchant) error {
	_, err := r.merchants.InsertOne(ctx, m)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *Repository) merchantByEmail(ctx context.Context, email string) (*Merchant, error) {
	var m Merchant
	err := r.merchants.FindOne(ctx, bson.M{"email": email}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrMerchantNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// countMerchants returns how many merchants exist (used to seed the demo on first boot only).
func (r *Repository) countMerchants(ctx context.Context) (int64, error) {
	return r.merchants.CountDocuments(ctx, bson.M{})
}

// upsertMerchant inserts the merchant if its email is free, otherwise returns the existing one. Used
// only by demo seeding, so it stays idempotent across restarts.
func (r *Repository) upsertMerchant(ctx context.Context, m *Merchant) (*Merchant, error) {
	existing, err := r.merchantByEmail(ctx, m.Email)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrMerchantNotFound) {
		return nil, err
	}
	if err := r.insertMerchant(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// upsertStore inserts or replaces a store by id (used by demo seeding).
func (r *Repository) upsertStore(ctx context.Context, s *Store) error {
	_, err := r.stores.ReplaceOne(ctx, bson.M{"_id": s.ID}, s, options.Replace().SetUpsert(true))
	return err
}

func (r *Repository) insertStore(ctx context.Context, s *Store) error {
	_, err := r.stores.InsertOne(ctx, s)
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicate
	}
	return err
}

func (r *Repository) storesByOwner(ctx context.Context, ownerID string) ([]Store, error) {
	cur, err := r.stores.Find(ctx, bson.M{"ownerId": ownerID})
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	var out []Store
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// allStores lists every store across all tenants (super-admin), newest first.
func (r *Repository) allStores(ctx context.Context) ([]Store, error) {
	cur, err := r.stores.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list all stores: %w", err)
	}
	var out []Store
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// allMerchants lists every merchant across the platform (super-admin), newest first.
func (r *Repository) allMerchants(ctx context.Context) ([]Merchant, error) {
	cur, err := r.merchants.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list all merchants: %w", err)
	}
	var out []Merchant
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// updatePlanByID flips any store's plan without an ownership guard (super-admin only).
func (r *Repository) updatePlanByID(ctx context.Context, id, tier string) (*Store, error) {
	var s Store
	err := r.stores.FindOneAndUpdate(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"plan": tier}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) storeByID(ctx context.Context, id string) (*Store, error) {
	var s Store
	err := r.stores.FindOne(ctx, bson.M{"_id": id}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// updatePlan flips a store's membership tier, guarded by ownership so a merchant can only change a
// plan for a store they own. The tier string is validated by the service before it reaches here.
func (r *Repository) updatePlan(ctx context.Context, id, ownerID, tier string) (*Store, error) {
	var s Store
	err := r.stores.FindOneAndUpdate(ctx,
		bson.M{"_id": id, "ownerId": ownerID},
		bson.M{"$set": bson.M{"plan": tier}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// updateDisabled flips a store's on/off switch, guarded by ownership.
func (r *Repository) updateDisabled(ctx context.Context, id, ownerID string, disabled bool) (*Store, error) {
	var s Store
	err := r.stores.FindOneAndUpdate(ctx,
		bson.M{"_id": id, "ownerId": ownerID},
		bson.M{"$set": bson.M{"disabled": disabled}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// deleteStore removes a store, guarded by ownership.
func (r *Repository) deleteStore(ctx context.Context, id, ownerID string) error {
	res, err := r.stores.DeleteOne(ctx, bson.M{"_id": id, "ownerId": ownerID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrStoreNotFound
	}
	return nil
}

func (r *Repository) storeBySlug(ctx context.Context, slug string) (*Store, error) {
	var s Store
	err := r.stores.FindOne(ctx, bson.M{"slug": slug}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// updateSettings patches a store's settings/displayName, but only when owned by ownerID (the guard
// prevents a merchant from editing someone else's store).
func (r *Repository) updateSettings(ctx context.Context, id, ownerID, displayName string, settings Settings) (*Store, error) {
	set := bson.M{"settings": settings}
	if displayName != "" {
		set["displayName"] = displayName
	}
	var s Store
	err := r.stores.FindOneAndUpdate(ctx,
		bson.M{"_id": id, "ownerId": ownerID},
		bson.M{"$set": set},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
