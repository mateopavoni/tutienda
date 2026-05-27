// Command orders runs the checkout saga across the inventory and catalog services.
package main

import (
	"context"
	"os"

	"github.com/mateopavoni/archive-commerce/internal/orders"
	"github.com/mateopavoni/archive-commerce/internal/platform/config"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/mongox"
)

func main() {
	log := httpx.NewLogger("orders")

	common := config.LoadCommon()
	addr := config.EnvString("ORDERS_ADDR", ":8083")
	inventoryURL := config.EnvString("INVENTORY_URL", "http://localhost:8082")
	catalogURL := config.EnvString("CATALOG_URL", "http://localhost:8081")

	ctx := context.Background()
	client, err := mongox.Connect(ctx, common.MongoURI)
	if err != nil {
		log.Error("mongo connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	repo := orders.NewRepository(client.Database("orders"))
	svc := orders.NewService(
		repo,
		orders.NewInventoryClient(inventoryURL),
		orders.NewCatalogClient(catalogURL),
		log,
	)

	handler := httpx.Chain(orders.NewHandler(svc).Routes(),
		httpx.Recover(log),
		httpx.Logger(log),
	)

	if err := httpx.Serve(addr, handler, log); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
