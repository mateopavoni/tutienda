package inventory

import (
	"context"

	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// defaultStock is the demo on-hand stock per SKU. SKUs match the variants seeded by the catalog
// service. AX1-85 is deliberately scarce so the no-oversell behaviour is visible under load.
var defaultStock = map[string]int{
	// Aero Strata_X1 (footwear)
	"AX1-7": 40, "AX1-75": 35, "AX1-8": 60, "AX1-85": 10, "AX1-9": 55, "AX1-95": 30, "AX1-10": 45, "AX1-105": 25,
	// System Jacket X-T (outerwear)
	"SJX-S": 20, "SJX-M": 30, "SJX-L": 30, "SJX-XL": 15,
	// Void Parka (outerwear)
	"VP-S": 12, "VP-M": 18, "VP-L": 14,
	// Utility Rig (accessories)
	"UR-OS": 80,
	// Aero-Tread Unit 01 (footwear)
	"ATU-8": 22, "ATU-9": 26, "ATU-10": 20,
	// Tactical Vest (apparel)
	"TV-M": 24, "TV-L": 28, "TV-XL": 16,
	// Utility Tote (accessories)
	"UT-OS": 120,
}

// SeedIfEmpty seeds stock only when no stock documents exist, so a restart with live data (and live
// reservations) is left untouched. Convenient for local development and the docker-compose demo.
func (s *Service) SeedIfEmpty(ctx context.Context) (int, error) {
	n, err := s.repo.countStocks(ctx)
	if err != nil {
		return 0, err
	}
	if n > 0 {
		return 0, nil
	}
	return s.SeedDefault(ctx)
}

// SeedDefault sets the on-hand amount for every demo SKU under the demo store. Idempotent: re-running
// resets availability but never touches reserved counts on existing docs.
func (s *Service) SeedDefault(ctx context.Context) (int, error) {
	for sku, qty := range defaultStock {
		if err := s.repo.upsertStock(ctx, tenant.DemoID, sku, qty); err != nil {
			return 0, err
		}
	}
	return len(defaultStock), nil
}
