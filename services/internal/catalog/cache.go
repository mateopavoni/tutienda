package catalog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// cacheGetJSON reads and decodes a cached value. It treats any miss or error as "not cached" so a
// flaky cache degrades to a database read instead of failing the request.
func cacheGetJSON[T any](ctx context.Context, c *redis.Client, key string) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}
	raw, err := c.Get(ctx, key).Bytes()
	if err != nil {
		return zero, false
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return zero, false
	}
	return v, true
}

// cacheSetJSON stores a value with a TTL, ignoring cache write errors (best-effort).
func cacheSetJSON(ctx context.Context, c *redis.Client, key string, v any, ttl time.Duration) {
	if c == nil {
		return
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return
	}
	_ = c.Set(ctx, key, raw, ttl).Err()
}

// cacheInvalidate drops a store's catalog keys. Called after a write so reads can't serve stale data.
// Scoped per store so one store's edit never evicts another store's cache.
func cacheInvalidate(ctx context.Context, c *redis.Client, tenant string) {
	if c == nil {
		return
	}
	keys, err := c.Keys(ctx, "catalog:"+tenant+":*").Result()
	if err != nil || len(keys) == 0 {
		return
	}
	_ = c.Del(ctx, keys...).Err()
}
