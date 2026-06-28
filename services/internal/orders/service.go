package orders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// CheckoutItem is what the client sends: just a SKU and a quantity. Price is never trusted from here.
type CheckoutItem struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

// Service runs the checkout saga across inventory and catalog.
type Service struct {
	repo    *Repository
	inv     *InventoryClient
	catalog *CatalogClient
	log     *slog.Logger
}

func NewService(repo *Repository, inv *InventoryClient, catalog *CatalogClient, log *slog.Logger) *Service {
	return &Service{repo: repo, inv: inv, catalog: catalog, log: log}
}

// Checkout is the saga:
//  1. resolve prices from the catalog (server-authoritative).
//  2. reserve stock — fail fast with OutOfStockError if any line is short.
//  3. persist a PENDING order holding the reservation id.
//  4. confirm the reservation (simulated payment); on failure, release it (compensation) and mark FAILED.
func (s *Service) Checkout(ctx context.Context, items []CheckoutItem) (*Order, error) {
	if len(items) == 0 {
		return nil, errors.New("cart is empty")
	}
	tid, ok := tenant.FromContext(ctx)
	if !ok {
		return nil, errors.New("missing tenant")
	}

	orderItems, lines, total, currency, err := s.price(ctx, items)
	if err != nil {
		return nil, err
	}

	reservationID, err := s.inv.Reserve(ctx, lines)
	if err != nil {
		return nil, err // OutOfStockError or transport error bubbles up; nothing persisted yet
	}

	// Optional buyer attribution: the gateway stamps X-Customer-ID only for a logged-in customer of this
	// store. A guest checkout leaves it empty — accounts are optional.
	customerID, _ := authx.CustomerFromContext(ctx)

	now := time.Now().UTC()
	order := &Order{
		TenantID:      tid,
		CustomerID:    customerID,
		Items:         orderItems,
		Currency:      currency,
		TotalCents:    total,
		Status:        StatusPending,
		ReservationID: reservationID,
		CreatedAt:     now,
	}
	id, err := s.repo.insert(ctx, order)
	if err != nil {
		// Don't leak the hold if we can't even record the order.
		if relErr := s.inv.Release(ctx, reservationID); relErr != nil {
			s.log.Error("release after insert failure failed", "error", relErr)
		}
		return nil, fmt.Errorf("persist order: %w", err)
	}
	order.ID = id

	if err := s.inv.Confirm(ctx, reservationID); err != nil {
		// Compensating action: give the stock back and record the failure.
		s.log.Error("payment confirm failed, compensating", "order", id.Hex(), "error", err)
		if relErr := s.inv.Release(ctx, reservationID); relErr != nil {
			s.log.Error("compensation release failed", "order", id.Hex(), "error", relErr)
		}
		_ = s.repo.setStatus(ctx, id, StatusFailed, err.Error())
		order.Status = StatusFailed
		order.FailureReason = err.Error()
		return order, nil
	}

	_ = s.repo.setStatus(ctx, id, StatusConfirmed, "")
	order.Status = StatusConfirmed
	return order, nil
}

// price resolves each item's name and price from the catalog and computes the total server-side.
func (s *Service) price(ctx context.Context, items []CheckoutItem) ([]Item, []reserveLine, int64, string, error) {
	orderItems := make([]Item, 0, len(items))
	lines := make([]reserveLine, 0, len(items))
	var total int64
	currency := ""

	for _, it := range items {
		if it.Qty <= 0 {
			return nil, nil, 0, "", fmt.Errorf("invalid quantity %d for %s", it.Qty, it.SKU)
		}
		v, err := s.catalog.Variant(ctx, it.SKU)
		if err != nil {
			if errors.Is(err, ErrVariantNotFound) {
				return nil, nil, 0, "", fmt.Errorf("unknown sku %q: %w", it.SKU, ErrVariantNotFound)
			}
			return nil, nil, 0, "", err
		}
		// Server-side drop gate: a scheduled drop can't be bought before its release time, even if a
		// client skips the storefront and calls the API directly. The no-oversell guard still applies
		// once it's live; this just enforces the "not yet" window.
		if v.DropAt != nil && time.Now().UTC().Before(*v.DropAt) {
			return nil, nil, 0, "", &DropNotReleasedError{SKU: it.SKU, DropAt: *v.DropAt}
		}
		if currency == "" {
			currency = v.Currency
		}
		name := v.ProductName
		if v.Label != "" {
			name = v.ProductName + " — " + v.Label
		}
		orderItems = append(orderItems, Item{SKU: it.SKU, Qty: it.Qty, Name: name, PriceCents: v.PriceCents})
		lines = append(lines, reserveLine{SKU: it.SKU, Qty: it.Qty})
		total += v.PriceCents * int64(it.Qty)
	}
	return orderItems, lines, total, currency, nil
}

// Get returns an order by id, scoped to the store.
func (s *Service) Get(ctx context.Context, tenant string, id primitive.ObjectID) (*Order, error) {
	return s.repo.get(ctx, tenant, id)
}

// Mine lists a customer's orders within their store, newest first (storefront order history).
func (s *Service) Mine(ctx context.Context, tenant, customerID string) ([]Order, error) {
	orders, err := s.repo.listByCustomer(ctx, tenant, customerID)
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []Order{}
	}
	return orders, nil
}
