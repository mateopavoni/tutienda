// Command inventory is the stock + reservations microservice — the concurrency core of the engine.
package main

import (
	"context"
	"os"
	"time"

	"github.com/mateopavoni/archive-commerce/internal/inventory"
	"github.com/mateopavoni/archive-commerce/internal/platform/config"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/mongox"
)

func main() {
	log := httpx.NewLogger("inventory")

	common := config.LoadCommon()
	addr := config.EnvString("INVENTORY_ADDR", ":8082")
	reservationTTL := config.EnvDuration("RESERVATION_TTL_SECONDS", 900)
	janitorInterval := config.EnvDuration("JANITOR_INTERVAL_SECONDS", 30)

	ctx := context.Background()
	client, err := mongox.Connect(ctx, common.MongoURI)
	if err != nil {
		log.Error("mongo connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	repo := inventory.NewRepository(client.Database("inventory"))
	indexCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := repo.EnsureIndexes(indexCtx); err != nil {
		log.Error("ensure indexes failed", "error", err)
		os.Exit(1)
	}

	svc := inventory.NewService(repo, reservationTTL, log)

	if n, err := svc.SeedIfEmpty(ctx); err != nil {
		log.Warn("seed-if-empty failed", "error", err)
	} else if n > 0 {
		log.Info("seeded inventory", "skus", n)
	}

	// Background janitor: returns stock held by expired reservations.
	janitorCtx, stopJanitor := context.WithCancel(ctx)
	defer stopJanitor()
	go inventory.NewJanitor(svc, janitorInterval, log).Run(janitorCtx)

	handler := httpx.Chain(inventory.NewHandler(svc).Routes(),
		httpx.Recover(log),
		httpx.Logger(log),
	)

	if err := httpx.Serve(addr, handler, log); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
