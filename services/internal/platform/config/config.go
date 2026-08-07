// Package config reads service configuration from environment variables.
// Keeping it dependency-free (no viper) makes the wiring obvious and easy to defend.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// DevJWTSecret is the well-known, publicly documented fallback signing secret used only in dev (this
// exact string is checked into the repo and README). A production deploy that still carries it would let
// anyone forge a JWT for any role, including the platform super-admin — see RequireJWTSecret.
const DevJWTSecret = "dev-insecure-secret-change-me"

// Common is configuration shared by every service.
type Common struct {
	MongoURI  string
	RedisAddr string
	// Env distinguishes a real deployment from local dev; only "prod" enables the strict checks below.
	// Defaults to "dev" so a bare `go run`/docker-compose-for-dev never needs to set it.
	Env string
	// JWTSecret signs/verifies admin session tokens. Services that validate tokens (accounts, gateway)
	// must share the same value and must call RequireJWTSecret right after LoadCommon; it is fine to
	// leave the dev default for the others (they never sign/verify a token).
	JWTSecret string
}

// IsProd reports whether the service is running in production mode (ENV=prod).
func (c Common) IsProd() bool { return c.Env == "prod" }

// RequireJWTSecret refuses to start the process (logs and os.Exit(1)) when running in production with an
// unset or still-default JWT_SECRET. Call this from any service that actually signs/verifies tokens
// (accounts, gateway) immediately after LoadCommon — services that never touch JWTSecret don't need it.
func (c Common) RequireJWTSecret(log *slog.Logger) {
	if c.IsProd() && c.JWTSecret == DevJWTSecret {
		log.Error("refusing to start: JWT_SECRET is unset or still the known dev default while ENV=prod " +
			"(anyone could forge a JWT for any role) — set a long random JWT_SECRET")
		os.Exit(1)
	}
}

func LoadCommon() Common {
	return Common{
		MongoURI:  env("MONGO_URI", "mongodb://localhost:27017"),
		RedisAddr: env("REDIS_ADDR", "localhost:6379"),
		Env:       env("ENV", "dev"),
		JWTSecret: env("JWT_SECRET", DevJWTSecret),
	}
}

// env returns the variable or a fallback when unset/empty.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// EnvString is exported for services that need ad-hoc lookups.
func EnvString(key, fallback string) string { return env(key, fallback) }

// EnvInt parses an integer env var, falling back on any parse error.
func EnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// EnvDuration reads an integer number of seconds into a time.Duration.
func EnvDuration(key string, fallbackSeconds int) time.Duration {
	return time.Duration(EnvInt(key, fallbackSeconds)) * time.Second
}

// EnvBool parses a boolean env var (accepts the strconv.ParseBool set: 1/t/T/TRUE/true/True,
// 0/f/F/FALSE/false/False), falling back on any parse error or unset var.
func EnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
