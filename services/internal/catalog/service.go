package catalog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// maxProductImages caps a product's gallery so a single document stays small (the storefront ships it on
// every catalog read).
const maxProductImages = 8

// ErrInvalidProduct is returned when an admin product payload is missing required fields.
var ErrInvalidProduct = errors.New("invalid product")

// ErrFeatureLocked is returned when a store's plan doesn't include a feature it tried to use (e.g. a
// free store scheduling a drop). The handler maps it to 402 Payment Required so the UI can prompt an upgrade.
var ErrFeatureLocked = errors.New("feature not in plan")

// ErrProductLimit is returned when a store has hit the product cap its plan allows.
var ErrProductLimit = errors.New("product limit reached for plan")

// Service serves products through a cache-aside layer: read Redis, fall back to Mongo on a miss,
// then populate Redis. Writes invalidate the cache so reads never serve stale data.
type Service struct {
	repo  *Repository
	cache *redis.Client
	ttl   time.Duration
	log   *slog.Logger
}

func NewService(repo *Repository, cache *redis.Client, ttl time.Duration, log *slog.Logger) *Service {
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return &Service{repo: repo, cache: cache, ttl: ttl, log: log}
}

// List returns a store's products for a category (empty = all), cached by tenant+category+limit.
func (s *Service) List(ctx context.Context, tenant, category string, limit int64) ([]Product, error) {
	key := fmt.Sprintf("catalog:%s:list:%s:%d", tenant, category, limit)
	if cached, ok := cacheGetJSON[[]Product](ctx, s.cache, key); ok {
		return cached, nil
	}
	products, err := s.repo.List(ctx, tenant, category, limit)
	if err != nil {
		return nil, err
	}
	if products == nil {
		products = []Product{}
	}
	cacheSetJSON(ctx, s.cache, key, products, s.ttl)
	return products, nil
}

// Get returns one of a store's products, cached by tenant+id.
func (s *Service) Get(ctx context.Context, tenant, id string) (*Product, error) {
	key := fmt.Sprintf("catalog:%s:product:%s", tenant, id)
	if cached, ok := cacheGetJSON[Product](ctx, s.cache, key); ok {
		return &cached, nil
	}
	p, err := s.repo.Get(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	cacheSetJSON(ctx, s.cache, key, *p, s.ttl)
	return p, nil
}

// Variant returns server-authoritative pricing for a store's SKU (used by the orders service).
func (s *Service) Variant(ctx context.Context, tenant, sku string) (*VariantInfo, error) {
	key := fmt.Sprintf("catalog:%s:variant:%s", tenant, sku)
	if cached, ok := cacheGetJSON[VariantInfo](ctx, s.cache, key); ok {
		return &cached, nil
	}
	v, err := s.repo.VariantBySKU(ctx, tenant, sku)
	if err != nil {
		return nil, err
	}
	cacheSetJSON(ctx, s.cache, key, *v, s.ttl)
	return v, nil
}

// Create inserts a new product for a store, assigning an id and creation time. Returns the stored
// product. Invalidates the store's cache.
func (s *Service) Create(ctx context.Context, tenant string, p Product) (*Product, error) {
	if p.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidProduct)
	}
	tier := plan.FromContext(ctx)
	if err := gateDrop(tier, p); err != nil {
		return nil, err
	}
	// Enforce the plan's product cap before inserting a new product.
	if limit := tier.ProductLimit(); limit != plan.Unlimited {
		count, err := s.repo.CountByTenant(ctx, tenant)
		if err != nil {
			return nil, err
		}
		if count >= int64(limit) {
			return nil, fmt.Errorf("%w: %s allows %d products", ErrProductLimit, tier, limit)
		}
	}
	p.TenantID = tenant
	if p.ID == "" {
		p.ID = primitive.NewObjectID().Hex()
	}
	if p.Currency == "" {
		p.Currency = "USD"
	}
	normalizeImages(&p)
	normalizeVariants(&p)
	p.CreatedAt = time.Now().UTC()
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, err
	}
	cacheInvalidate(ctx, s.cache, tenant)
	return &p, nil
}

// normalizeImages builds the gallery on write as the unique, non-blank list of [ImageURL, ...Images],
// caps it, and pins ImageURL = Images[0]. So the card/seed (which know only ImageURL) and the detail
// gallery (which renders Images) can never disagree about the primary image. Defensive write-time stance,
// same as the layout sanitizers.
func normalizeImages(p *Product) {
	out := make([]string, 0, len(p.Images)+1)
	seen := map[string]bool{}
	add := func(im string) {
		im = strings.TrimSpace(im)
		if im == "" || seen[im] || len(out) >= maxProductImages {
			return
		}
		seen[im] = true
		out = append(out, im)
	}
	add(p.ImageURL) // primary first
	for _, im := range p.Images {
		add(im)
	}
	p.Images = out
	if len(out) > 0 {
		p.ImageURL = out[0]
	}
}

// normalizeVariants cleans the variant list on write: trims fields, drops blank-SKU rows and duplicate
// SKUs (inventory keys on {tenantId, sku}, so a repeated SKU would make two variants fight over one stock
// document), and falls back the label to the SKU. Same defensive write-time stance as normalizeImages and
// the layout sanitizers — a multi-variant product (sizes/colors) can't be saved into an inconsistent state.
func normalizeVariants(p *Product) {
	out := make([]Variant, 0, len(p.Variants))
	seen := map[string]bool{}
	for _, v := range p.Variants {
		v.SKU = strings.TrimSpace(v.SKU)
		v.Label = strings.TrimSpace(v.Label)
		v.Color = strings.TrimSpace(v.Color)
		if v.SKU == "" || seen[v.SKU] {
			continue
		}
		if v.Label == "" {
			v.Label = v.SKU
		}
		seen[v.SKU] = true
		out = append(out, v)
	}
	p.Variants = out
}

// Update replaces an existing product within a store, preserving its creation time. Invalidates cache.
func (s *Service) Update(ctx context.Context, tenant, id string, p Product) (*Product, error) {
	if err := gateDrop(plan.FromContext(ctx), p); err != nil {
		return nil, err
	}
	existing, err := s.repo.Get(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	p.ID = id
	p.TenantID = tenant
	p.CreatedAt = existing.CreatedAt
	if p.Currency == "" {
		p.Currency = existing.Currency
	}
	normalizeImages(&p)
	normalizeVariants(&p)
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, err
	}
	cacheInvalidate(ctx, s.cache, tenant)
	return &p, nil
}

// Delete removes a store's product and invalidates the cache.
func (s *Service) Delete(ctx context.Context, tenant, id string) error {
	if err := s.repo.Delete(ctx, tenant, id); err != nil {
		return err
	}
	cacheInvalidate(ctx, s.cache, tenant)
	return nil
}

// gateDrop blocks scheduling a drop (any dropAt) on a plan that doesn't unlock the drops tool. The
// storefront and the orders saga already gate buying before release; this stops a free store from
// authoring a drop in the first place — the server is the authority, not the dimmed UI control.
func gateDrop(tier plan.Tier, p Product) error {
	if p.DropAt != nil && !tier.Has(plan.FeatureDrops) {
		return fmt.Errorf("%w: drops require an upgrade", ErrFeatureLocked)
	}
	return nil
}

// SeedIfEmpty seeds the catalog only when it has no products, so a restart with existing data is a
// no-op. Convenient for local development and the docker-compose demo.
func (s *Service) SeedIfEmpty(ctx context.Context) (int, error) {
	n, err := s.repo.Count(ctx)
	if err != nil {
		return 0, err
	}
	if n > 0 {
		return 0, nil
	}
	return s.Seed(ctx)
}

// Seed populates the demo store's catalog and clears its cache.
func (s *Service) Seed(ctx context.Context) (int, error) {
	for _, p := range demoProducts() {
		p.TenantID = tenant.DemoID
		if err := s.repo.Upsert(ctx, p); err != nil {
			return 0, err
		}
	}
	cacheInvalidate(ctx, s.cache, tenant.DemoID)
	return len(demoProducts()), nil
}
