package inventory

import (
	"context"

	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// demoStock is the per-store on-hand stock, keyed by slug then SKU. SKUs match the variants the catalog
// service seeds for the same store. In SYSTEM ARCHIVE, AX1-85 is deliberately scarce so the no-oversell
// behaviour is visible under load (the loadtest target). See tenant.DemoStores for the example stores.
var demoStock = map[string]map[string]int{
	tenant.DemoSlug: {
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
	},
	"casa-bruta": {
		"CB-LAMP": 18, "CB-STOOL-BLK": 22, "CB-STOOL-GRY": 22, "CB-THROW-SND": 40, "CB-THROW-SLT": 40,
		"CB-VES-S": 30, "CB-VES-M": 26, "CB-VES-L": 14, "CB-SHELF": 25,
	},
	"papel-y-tinta": {
		"PT-ARCH": 35, "PT-POEMS": 60, "PT-ESSAYS": 50, "PT-GRID": 45, "PT-TYPE": 20,
	},
	"cafe-noventa": {
		"CN-ETH-250": 80, "CN-ETH-1K": 30, "CN-ESP-250": 90, "CN-ESP-1K": 35, "CN-COLD": 40,
		"CN-DRIP": 25, "CN-MUG": 60,
	},
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

// SeedDefault sets the on-hand amount for every demo SKU under each example store. Idempotent:
// re-running resets availability but never touches reserved counts on existing docs.
func (s *Service) SeedDefault(ctx context.Context) (int, error) {
	total := 0
	for _, ds := range tenant.DemoStores {
		for sku, qty := range demoStock[ds.Slug] {
			if err := s.repo.upsertStock(ctx, ds.ID, sku, qty); err != nil {
				return 0, err
			}
			total++
		}
	}
	return total, nil
}
