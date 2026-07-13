// Command accounts is the identity + stores service: merchant signup/login (issuing session tokens)
// and the per-store (tenant) registry the rest of the fleet is scoped against.
package main

import (
	"context"
	"os"

	"github.com/mateopavoni/archive-commerce/internal/accounts"
	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/config"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/mongox"
)

func main() {
	log := httpx.NewLogger("accounts")

	common := config.LoadCommon()
	addr := config.EnvString("ACCOUNTS_ADDR", ":8084")
	tokenTTL := config.EnvDuration("ACCOUNTS_TOKEN_TTL_SECONDS", 24*60*60)
	maxStores := config.EnvInt("MAX_STORES_PER_MERCHANT", 2) // demo anti-abuse; 0 = unlimited
	catalogURL := config.EnvString("CATALOG_URL", "http://localhost:8081")
	inventoryURL := config.EnvString("INVENTORY_URL", "http://localhost:8082")

	ctx := context.Background()
	client, err := mongox.Connect(ctx, common.MongoURI)
	if err != nil {
		log.Error("mongo connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	repo := accounts.NewRepository(client.Database("accounts"))
	if err := repo.EnsureIndexes(ctx); err != nil {
		log.Error("ensure indexes failed", "error", err)
		os.Exit(1)
	}

	issuer := authx.NewIssuer(common.JWTSecret, tokenTTL)
	svc := accounts.NewService(repo, issuer, log, maxStores,
		accounts.NewCatalogClient(catalogURL), accounts.NewInventoryClient(inventoryURL))

	if seeded, err := svc.SeedDemoIfEmpty(ctx); err != nil {
		log.Warn("seed-demo failed", "error", err)
	} else if seeded {
		log.Info("seeded demo store", "slug", "system-archive")
	}

	handler := httpx.Chain(accounts.NewHandler(svc).Routes(),
		httpx.Recover(log),
		httpx.Logger(log),
	)

	if err := httpx.Serve(addr, handler, log); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
