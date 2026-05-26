// Command catalog is the read-heavy product service, served cache-aside through Redis.
package main

import (
	"context"
	"os"

	"github.com/mateopavoni/archive-commerce/internal/catalog"
	"github.com/mateopavoni/archive-commerce/internal/platform/config"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/media"
	"github.com/mateopavoni/archive-commerce/internal/platform/mongox"
	"github.com/mateopavoni/archive-commerce/internal/platform/redisx"
)

func main() {
	log := httpx.NewLogger("catalog")

	common := config.LoadCommon()
	addr := config.EnvString("CATALOG_ADDR", ":8081")
	cacheTTL := config.EnvDuration("CATALOG_CACHE_TTL_SECONDS", 60)

	ctx := context.Background()
	client, err := mongox.Connect(ctx, common.MongoURI)
	if err != nil {
		log.Error("mongo connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	cache, err := redisx.Connect(ctx, common.RedisAddr)
	if err != nil {
		log.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = cache.Close() }()

	repo := catalog.NewRepository(client.Database("catalog"))
	svc := catalog.NewService(repo, cache, cacheTTL, log)

	// Object store for merchant-uploaded images. Optional: with no MINIO_ENDPOINT the client is nil and
	// the upload endpoint returns 503, so the service still boots in environments without object storage.
	mediaClient, err := media.New(ctx, media.Config{
		Endpoint:   config.EnvString("MINIO_ENDPOINT", ""),
		AccessKey:  config.EnvString("MINIO_ACCESS_KEY", ""),
		SecretKey:  config.EnvString("MINIO_SECRET_KEY", ""),
		Bucket:     config.EnvString("MINIO_BUCKET", "media"),
		PublicBase: config.EnvString("PUBLIC_MEDIA_BASE", ""),
		UseSSL:     config.EnvInt("MINIO_USE_SSL", 0) == 1,
	})
	if err != nil {
		log.Warn("media storage init failed, uploads disabled", "error", err)
	} else if mediaClient != nil {
		log.Info("media storage ready")
	}

	if n, err := svc.SeedIfEmpty(ctx); err != nil {
		log.Warn("seed-if-empty failed", "error", err)
	} else if n > 0 {
		log.Info("seeded catalog", "products", n)
	}

	handler := httpx.Chain(catalog.NewHandler(svc, mediaClient).Routes(),
		httpx.Recover(log),
		httpx.Logger(log),
	)

	if err := httpx.Serve(addr, handler, log); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
