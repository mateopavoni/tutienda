// Command gateway is the single public entry point: rate limiting, CORS and routing to services.
package main

import (
	"context"
	"os"

	"github.com/mateopavoni/archive-commerce/internal/gateway"
	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/config"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/redisx"
)

func main() {
	log := httpx.NewLogger("gateway")

	common := config.LoadCommon()
	common.RequireJWTSecret(log) // gateway verifies store/admin/customer JWTs; never boot in prod with the dev default
	addr := config.EnvString("GATEWAY_ADDR", ":8080")
	webOrigin := config.EnvString("WEB_ORIGIN", "http://localhost:5173")
	rateMax := config.EnvInt("RATE_LIMIT_MAX", 100)
	rateWindow := config.EnvDuration("RATE_LIMIT_WINDOW_SECONDS", 60)
	accountsURL := config.EnvString("ACCOUNTS_URL", "http://localhost:8084")
	tenantCacheTTL := config.EnvDuration("TENANT_CACHE_TTL_SECONDS", 300)
	tokenTTL := config.EnvDuration("ACCOUNTS_TOKEN_TTL_SECONDS", 24*60*60)

	ctx := context.Background()
	redisClient, err := redisx.Connect(ctx, common.RedisAddr)
	if err != nil {
		log.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = redisClient.Close() }()

	router, err := gateway.Router(gateway.Config{
		CatalogURL:   config.EnvString("CATALOG_URL", "http://localhost:8081"),
		InventoryURL: config.EnvString("INVENTORY_URL", "http://localhost:8082"),
		OrdersURL:    config.EnvString("ORDERS_URL", "http://localhost:8083"),
		AccountsURL:  accountsURL,
		Resolver:     gateway.NewResolver(accountsURL, redisClient, tenantCacheTTL),
		Issuer:       authx.NewIssuer(common.JWTSecret, tokenTTL),
		Limiter:      gateway.NewRateLimiter(redisClient, rateMax, rateWindow),
		Log:          log,
	})
	if err != nil {
		log.Error("router build failed", "error", err)
		os.Exit(1)
	}

	handler := httpx.Chain(router,
		httpx.Recover(log),
		httpx.Logger(log),
		httpx.SecurityHeaders(),
		httpx.CORS(webOrigin),
	)

	if err := httpx.Serve(addr, handler, log); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
