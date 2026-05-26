package catalog

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrProductNotFound is returned when a product id (or SKU) has no match.
var ErrProductNotFound = errors.New("product not found")

// Repository is the MongoDB persistence layer for products.
type Repository struct {
	products *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{products: db.Collection("products")}
}

// List returns a store's products, optionally filtered by category, newest first.
func (r *Repository) List(ctx context.Context, tenant, category string, limit int64) ([]Product, error) {
	filter := bson.M{"tenantId": tenant}
	if category != "" {
		filter["category"] = category
	}
	cur, err := r.products.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	var out []Product
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get returns one of a store's products by id.
func (r *Repository) Get(ctx context.Context, tenant, id string) (*Product, error) {
	var p Product
	err := r.products.FindOne(ctx, bson.M{"_id": id, "tenantId": tenant}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// VariantBySKU finds the product owning a SKU within a store and returns the pricing info for it.
func (r *Repository) VariantBySKU(ctx context.Context, tenant, sku string) (*VariantInfo, error) {
	var p Product
	err := r.products.FindOne(ctx, bson.M{"tenantId": tenant, "variants.sku": sku}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	for _, v := range p.Variants {
		if v.SKU == sku {
			return &VariantInfo{
				SKU:         v.SKU,
				ProductID:   p.ID,
				ProductName: p.Name,
				Label:       v.Label,
				PriceCents:  p.PriceCents,
				Currency:    p.Currency,
				DropAt:      p.DropAt,
			}, nil
		}
	}
	return nil, ErrProductNotFound
}

// Count returns how many products exist across all stores (used to seed on first boot only).
func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.products.CountDocuments(ctx, bson.M{})
}

// CountByTenant returns how many products a single store holds (used to enforce the plan's product cap).
func (r *Repository) CountByTenant(ctx context.Context, tenant string) (int64, error) {
	return r.products.CountDocuments(ctx, bson.M{"tenantId": tenant})
}

// Upsert inserts or replaces a product, scoped to its store (used by seeding and the admin editor).
func (r *Repository) Upsert(ctx context.Context, p Product) error {
	_, err := r.products.ReplaceOne(ctx, bson.M{"_id": p.ID, "tenantId": p.TenantID}, p, options.Replace().SetUpsert(true))
	return err
}

// Delete removes one of a store's products.
func (r *Repository) Delete(ctx context.Context, tenant, id string) error {
	res, err := r.products.DeleteOne(ctx, bson.M{"_id": id, "tenantId": tenant})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrProductNotFound
	}
	return nil
}
