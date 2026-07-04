package accounts

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// Demo credentials for the seeded showcase. The merchant owns the SYSTEM ARCHIVE store; the admin is
// the SaaS super-admin (RoleAdmin) used to reach the cross-tenant console at /admin. Documented so both
// are loginable out of the box; override in any real deployment.
const (
	demoEmail    = "demo@system-archive.store"
	demoPassword = "demo-archive-2026"

	adminEmail    = "admin@tutienda.store"
	adminPassword = "admin-tutienda-2026"
)

// SeedDemoIfEmpty creates the demo merchant and the "SYSTEM ARCHIVE" store (under tenant.DemoID) when
// no merchants exist yet, so the showcase storefront resolves and the demo is manageable out of the
// box. Idempotent and a no-op once any merchant exists.
func (s *Service) SeedDemoIfEmpty(ctx context.Context) (bool, error) {
	n, err := s.repo.countMerchants(ctx)
	if err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}

	hash, err := authx.HashPassword(demoPassword)
	if err != nil {
		return false, err
	}
	merchant := &Merchant{
		ID:           primitive.NewObjectID(),
		Email:        demoEmail,
		PasswordHash: hash,
		Role:         authx.RoleMerchant,
		CreatedAt:    time.Now().UTC(),
	}
	merchant, err = s.repo.upsertMerchant(ctx, merchant)
	if err != nil {
		return false, err
	}

	// Seed the platform super-admin so the cross-tenant console at /admin is reachable out of the box.
	adminHash, err := authx.HashPassword(adminPassword)
	if err != nil {
		return false, err
	}
	if _, err := s.repo.upsertMerchant(ctx, &Merchant{
		ID:           primitive.NewObjectID(),
		Email:        adminEmail,
		PasswordHash: adminHash,
		Role:         authx.RoleAdmin,
		CreatedAt:    time.Now().UTC(),
	}); err != nil {
		return false, err
	}

	// Seed every example store (see tenant.DemoStores) under the demo merchant, so one login shows the
	// whole multi-tenant spread in /app and each rubro's storefront resolves out of the box. Each gets a
	// full-bleed hero (with a rubro-appropriate photo) + an about block over the mandatory product grid
	// so the homepage looks built, not empty.
	now := time.Now().UTC()
	for _, ds := range tenant.DemoStores {
		store := &Store{
			ID:          ds.ID,
			Slug:        ds.Slug,
			OwnerID:     merchant.ID.Hex(),
			DisplayName: ds.Name,
			Settings: Settings{
				AccentColor: ds.Accent,
				Currency:    "USD",
				Theme:       normalizeTheme(ds.Theme),
				Layout: []Section{
					{Type: "hero", Enabled: true, Heading: ds.Name, Body: ds.Tagline, ImageURL: ds.HeroImage, CTALabel: "Shop now", Overlay: heroOverlayDefault, OverlayColor: "black"},
					{Type: "catalog", Enabled: true},
					{Type: "about", Enabled: true, Heading: "About", Body: ds.About},
				},
			},
			Plan:      string(plan.Normalize(ds.Plan)), // unknown/typo tier ⇒ free, same defensive stance as everywhere
			CreatedAt: now,
		}
		if err := s.repo.upsertStore(ctx, store); err != nil {
			return false, err
		}
	}
	return true, nil
}
