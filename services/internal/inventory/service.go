package inventory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ErrNotHeld means the reservation exists but is no longer in the HELD state, so it can't be
// confirmed or released again (it was already finalized, possibly by the janitor).
var ErrNotHeld = errors.New("reservation is not in HELD state")

// OutOfStockError reports which SKU could not be reserved.
type OutOfStockError struct{ SKU string }

func (e *OutOfStockError) Error() string { return "out of stock: " + e.SKU }

// Service holds the inventory business rules. It is the only place that decides how stock moves.
type Service struct {
	repo       *Repository
	defaultTTL time.Duration
	log        *slog.Logger
}

func NewService(repo *Repository, defaultTTL time.Duration, log *slog.Logger) *Service {
	return &Service{repo: repo, defaultTTL: defaultTTL, log: log}
}

// Reserve holds stock for every line atomically per SKU, scoped to one store. A multi-SKU reservation
// is not globally atomic, so if a later line is out of stock the earlier successful holds are
// compensated (released) before returning. The result is all-or-nothing from the caller's point of view.
func (s *Service) Reserve(ctx context.Context, tenant string, lines []Line, ttl time.Duration) (*Reservation, error) {
	if len(lines) == 0 {
		return nil, errors.New("reservation needs at least one line")
	}
	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	done := make([]Line, 0, len(lines))
	for _, l := range lines {
		if l.Qty <= 0 {
			s.compensate(ctx, tenant, done)
			return nil, fmt.Errorf("invalid quantity %d for %s", l.Qty, l.SKU)
		}
		ok, err := s.repo.TryReserve(ctx, tenant, l.SKU, l.Qty)
		if err != nil {
			s.compensate(ctx, tenant, done)
			return nil, err
		}
		if !ok {
			s.compensate(ctx, tenant, done)
			return nil, &OutOfStockError{SKU: l.SKU}
		}
		done = append(done, l)
	}

	now := time.Now().UTC()
	res := &Reservation{
		TenantID:  tenant,
		Lines:     lines,
		Status:    StatusHeld,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
	id, err := s.repo.insertReservation(ctx, res)
	if err != nil {
		s.compensate(ctx, tenant, done)
		return nil, fmt.Errorf("persist reservation: %w", err)
	}
	res.ID = id
	return res, nil
}

// compensate releases stock that was already held when a later step fails.
func (s *Service) compensate(ctx context.Context, tenant string, lines []Line) {
	for _, l := range lines {
		if err := s.repo.release(ctx, tenant, l.SKU, l.Qty); err != nil {
			// Best-effort: log and move on. Stock for this SKU may need reconciliation.
			s.log.Error("compensation release failed", "tenant", tenant, "sku", l.SKU, "qty", l.Qty, "error", err)
		}
	}
}

// Confirm finalizes a sale. The atomic HELD→CONFIRMED claim guarantees only one actor finalizes
// the reservation, so stock is consumed exactly once even if the janitor races a user confirm. The
// claim is scoped to the store so a reservation can only be confirmed by its own tenant.
func (s *Service) Confirm(ctx context.Context, tenant string, id primitive.ObjectID) (*Reservation, error) {
	res, err := s.repo.claim(ctx, id, tenant, StatusHeld, StatusConfirmed)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, s.explainUnclaimable(ctx, id)
	}
	for _, l := range res.Lines {
		if err := s.repo.consume(ctx, tenant, l.SKU, l.Qty); err != nil {
			return nil, fmt.Errorf("consume %s: %w", l.SKU, err)
		}
	}
	res.Status = StatusConfirmed
	return res, nil
}

// Release returns held stock to available. Like Confirm, it is gated by the atomic claim and scoped
// to the store.
func (s *Service) Release(ctx context.Context, tenant string, id primitive.ObjectID) (*Reservation, error) {
	res, err := s.repo.claim(ctx, id, tenant, StatusHeld, StatusReleased)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, s.explainUnclaimable(ctx, id)
	}
	for _, l := range res.Lines {
		if err := s.repo.release(ctx, tenant, l.SKU, l.Qty); err != nil {
			return nil, fmt.Errorf("release %s: %w", l.SKU, err)
		}
	}
	res.Status = StatusReleased
	return res, nil
}

// explainUnclaimable tells apart "missing" from "already finalized" for a clearer error.
func (s *Service) explainUnclaimable(ctx context.Context, id primitive.ObjectID) error {
	res, err := s.repo.getReservation(ctx, id)
	if err != nil {
		return err // ErrReservationNotFound
	}
	return fmt.Errorf("%w (current status %s)", ErrNotHeld, res.Status)
}

// ReleaseExpired is what the janitor calls each tick: find HELD reservations past their expiry,
// claim each one, return its stock, and delete it. Reservations confirmed in the meantime are
// skipped because the claim loses the race. Returns how many were released.
func (s *Service) ReleaseExpired(ctx context.Context, batch int64) (int, error) {
	expired, err := s.repo.findExpiredHeld(ctx, time.Now().UTC(), batch)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, r := range expired {
		// Janitor sweeps every store; it claims by id ("" tenant) since it already holds the document,
		// then compensates stock against that reservation's own store.
		claimed, err := s.repo.claim(ctx, r.ID, "", StatusHeld, StatusReleased)
		if err != nil {
			s.log.Error("janitor claim failed", "reservation", r.ID.Hex(), "error", err)
			continue
		}
		if claimed == nil {
			continue // confirmed by a user between the find and the claim
		}
		s.compensate(ctx, claimed.TenantID, claimed.Lines)
		if err := s.repo.deleteReservation(ctx, r.ID); err != nil {
			s.log.Error("janitor delete failed", "reservation", r.ID.Hex(), "error", err)
		}
		released++
	}
	return released, nil
}

// Stocks returns current stock for the given SKUs within a store (storefront uses this to show
// availability).
func (s *Service) Stocks(ctx context.Context, tenant string, skus []string) ([]Stock, error) {
	return s.repo.getStocks(ctx, tenant, skus)
}

// SetStock sets the on-hand amount for a store's SKU (admin stock editor).
func (s *Service) SetStock(ctx context.Context, tenant, sku string, available int) error {
	if sku == "" {
		return errors.New("sku is required")
	}
	if available < 0 {
		return errors.New("available must be >= 0")
	}
	return s.repo.upsertStock(ctx, tenant, sku, available)
}

// DeleteTenant removes every stock and reservation of a store, called by accounts when the store itself
// is deleted (cascade). Idempotent: an already-empty tenant is not an error.
func (s *Service) DeleteTenant(ctx context.Context, tenant string) error {
	return s.repo.deleteByTenant(ctx, tenant)
}
