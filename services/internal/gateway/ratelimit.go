package gateway

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter is a distributed sliding-window limiter backed by a Redis sorted set per client.
// Each request is a member scored by its timestamp; expired members are trimmed on every check, so
// the window truly slides instead of resetting on a fixed boundary. Because the state lives in
// Redis, N gateway replicas share one limit.
type RateLimiter struct {
	redis  *redis.Client
	max    int
	window time.Duration
}

func NewRateLimiter(client *redis.Client, max int, window time.Duration) *RateLimiter {
	if max <= 0 {
		max = 100
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{redis: client, max: max, window: window}
}

// Allow records the request and reports whether the client is within its limit for the window.
// On any Redis error it fails open (allows the request) rather than locking users out.
func (rl *RateLimiter) Allow(ctx context.Context, clientID string) (bool, error) {
	key := "rate:" + clientID
	now := time.Now().UnixNano()
	windowStart := now - rl.window.Nanoseconds()
	member := strconv.FormatInt(now, 10) + "-" + strconv.Itoa(rand.Intn(1_000_000))

	pipe := rl.redis.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
	count := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return true, err // fail open
	}
	return count.Val() <= int64(rl.max), nil
}

// Limit returns the configured maximum (used to set response headers).
func (rl *RateLimiter) Limit() int { return rl.max }

// Window returns the configured window (used for Retry-After).
func (rl *RateLimiter) Window() time.Duration { return rl.window }
