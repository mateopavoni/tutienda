package inventory

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mateopavoni/archive-commerce/internal/platform/mongox"
)

// tnt is the store every single-tenant test runs against.
const tnt = "tenant-a"

// newTestEnv spins up a service backed by a throwaway database. It skips when MONGO_URI is unset,
// so the suite is a no-op without a database (run it via `docker compose run --rm backend-test`).
func newTestEnv(t *testing.T) (*Service, *Repository, func()) {
	t.Helper()
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		t.Skip("MONGO_URI not set; skipping integration test")
	}

	ctx := context.Background()
	client, err := mongox.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	dbName := "inv_test_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	db := client.Database(dbName)
	repo := NewRepository(db)
	if err := repo.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewService(repo, time.Minute, log)

	cleanup := func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	}
	return svc, repo, cleanup
}

// TestReserve_NoOversell is the central guarantee: when far more requests than units arrive at once,
// exactly the available count succeed, stock never goes negative, and held == sold-out amount.
func TestReserve_NoOversell(t *testing.T) {
	svc, repo, cleanup := newTestEnv(t)
	defer cleanup()
	ctx := context.Background()

	const stock = 10
	const concurrentBuyers = 250
	if err := repo.upsertStock(ctx, tnt, "AX1-85", stock); err != nil {
		t.Fatalf("seed stock: %v", err)
	}

	var success int64
	var wg sync.WaitGroup
	for i := 0; i < concurrentBuyers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.Reserve(ctx, tnt, []Line{{SKU: "AX1-85", Qty: 1}}, time.Minute); err == nil {
				atomic.AddInt64(&success, 1)
			}
		}()
	}
	wg.Wait()

	if success != stock {
		t.Fatalf("expected exactly %d successful reservations, got %d", stock, success)
	}
	stocks, err := svc.Stocks(ctx, tnt, []string{"AX1-85"})
	if err != nil || len(stocks) != 1 {
		t.Fatalf("read stock: %v (len %d)", err, len(stocks))
	}
	if stocks[0].Available != 0 {
		t.Fatalf("available should be 0 after selling out, got %d", stocks[0].Available)
	}
	if stocks[0].Available < 0 {
		t.Fatalf("oversold: available went negative (%d)", stocks[0].Available)
	}
	if stocks[0].Reserved != stock {
		t.Fatalf("reserved should be %d, got %d", stock, stocks[0].Reserved)
	}
}

// TestNoOversell_TenantIsolation is the multi-tenant guarantee: two stores hold the same SKU string
// with independent stock. Concurrent buyers hammering both stores sell out each one exactly to its own
// count, and neither store's load can consume the other's inventory.
func TestNoOversell_TenantIsolation(t *testing.T) {
	svc, repo, cleanup := newTestEnv(t)
	defer cleanup()
	ctx := context.Background()

	const (
		sku     = "SHARED-SKU"
		stockA  = 8
		stockB  = 3
		buyers  = 200
	)
	if err := repo.upsertStock(ctx, "store-a", sku, stockA); err != nil {
		t.Fatalf("seed store-a: %v", err)
	}
	if err := repo.upsertStock(ctx, "store-b", sku, stockB); err != nil {
		t.Fatalf("seed store-b: %v", err)
	}

	var successA, successB int64
	var wg sync.WaitGroup
	for i := 0; i < buyers; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := svc.Reserve(ctx, "store-a", []Line{{SKU: sku, Qty: 1}}, time.Minute); err == nil {
				atomic.AddInt64(&successA, 1)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err := svc.Reserve(ctx, "store-b", []Line{{SKU: sku, Qty: 1}}, time.Minute); err == nil {
				atomic.AddInt64(&successB, 1)
			}
		}()
	}
	wg.Wait()

	if successA != stockA {
		t.Fatalf("store-a: expected exactly %d successes, got %d", stockA, successA)
	}
	if successB != stockB {
		t.Fatalf("store-b: expected exactly %d successes, got %d", stockB, successB)
	}
	sa, _ := svc.Stocks(ctx, "store-a", []string{sku})
	sb, _ := svc.Stocks(ctx, "store-b", []string{sku})
	if sa[0].Available != 0 || sa[0].Reserved != stockA {
		t.Fatalf("store-a final stock wrong: available=%d reserved=%d", sa[0].Available, sa[0].Reserved)
	}
	if sb[0].Available != 0 || sb[0].Reserved != stockB {
		t.Fatalf("store-b final stock wrong: available=%d reserved=%d", sb[0].Available, sb[0].Reserved)
	}
}

// TestConfirmConsumesStock checks a confirmed reservation removes goods for good.
func TestConfirmConsumesStock(t *testing.T) {
	svc, repo, cleanup := newTestEnv(t)
	defer cleanup()
	ctx := context.Background()
	_ = repo.upsertStock(ctx, tnt, "SJX-M", 5)

	res, err := svc.Reserve(ctx, tnt, []Line{{SKU: "SJX-M", Qty: 2}}, time.Minute)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if _, err := svc.Confirm(ctx, tnt, res.ID); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	stocks, _ := svc.Stocks(ctx, tnt, []string{"SJX-M"})
	if stocks[0].Available != 3 || stocks[0].Reserved != 0 {
		t.Fatalf("after confirm want available=3 reserved=0, got available=%d reserved=%d", stocks[0].Available, stocks[0].Reserved)
	}
	// Confirming twice must not double-consume.
	if _, err := svc.Confirm(ctx, tnt, res.ID); err == nil {
		t.Fatalf("second confirm should fail (not HELD)")
	}
}

// TestReleaseRestoresStock checks releasing a hold returns the units to available.
func TestReleaseRestoresStock(t *testing.T) {
	svc, repo, cleanup := newTestEnv(t)
	defer cleanup()
	ctx := context.Background()
	_ = repo.upsertStock(ctx, tnt, "VP-L", 4)

	res, err := svc.Reserve(ctx, tnt, []Line{{SKU: "VP-L", Qty: 3}}, time.Minute)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if _, err := svc.Release(ctx, tnt, res.ID); err != nil {
		t.Fatalf("release: %v", err)
	}
	stocks, _ := svc.Stocks(ctx, tnt, []string{"VP-L"})
	if stocks[0].Available != 4 || stocks[0].Reserved != 0 {
		t.Fatalf("after release want available=4 reserved=0, got available=%d reserved=%d", stocks[0].Available, stocks[0].Reserved)
	}
}

// TestMultiSKUCompensation checks that when one line in a multi-SKU reservation is out of stock,
// the lines already held are released — the reservation is all-or-nothing.
func TestMultiSKUCompensation(t *testing.T) {
	svc, repo, cleanup := newTestEnv(t)
	defer cleanup()
	ctx := context.Background()
	_ = repo.upsertStock(ctx, tnt, "AX1-7", 5) // plenty
	_ = repo.upsertStock(ctx, tnt, "AX1-8", 1) // scarce

	_, err := svc.Reserve(ctx, tnt, []Line{{SKU: "AX1-7", Qty: 5}, {SKU: "AX1-8", Qty: 2}}, time.Minute)
	var oos *OutOfStockError
	if !errors.As(err, &oos) || oos.SKU != "AX1-8" {
		t.Fatalf("expected out-of-stock on AX1-8, got %v", err)
	}
	stocks, _ := svc.Stocks(ctx, tnt, []string{"AX1-7"})
	if stocks[0].Available != 5 || stocks[0].Reserved != 0 {
		t.Fatalf("AX1-7 should be fully restored after compensation, got available=%d reserved=%d", stocks[0].Available, stocks[0].Reserved)
	}
}

// TestJanitorReleasesExpired checks expired holds return their stock when the janitor sweeps.
func TestJanitorReleasesExpired(t *testing.T) {
	svc, repo, cleanup := newTestEnv(t)
	defer cleanup()
	ctx := context.Background()
	_ = repo.upsertStock(ctx, tnt, "UT-OS", 10)

	if _, err := svc.Reserve(ctx, tnt, []Line{{SKU: "UT-OS", Qty: 4}}, 5*time.Millisecond); err != nil {
		t.Fatalf("reserve: %v", err)
	}
	time.Sleep(30 * time.Millisecond)

	released, err := svc.ReleaseExpired(ctx, 100)
	if err != nil {
		t.Fatalf("release expired: %v", err)
	}
	if released != 1 {
		t.Fatalf("expected 1 expired reservation released, got %d", released)
	}
	stocks, _ := svc.Stocks(ctx, tnt, []string{"UT-OS"})
	if stocks[0].Available != 10 || stocks[0].Reserved != 0 {
		t.Fatalf("janitor should restore stock, got available=%d reserved=%d", stocks[0].Available, stocks[0].Reserved)
	}
}
