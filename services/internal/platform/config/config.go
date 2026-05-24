// Package config reads service configuration from environment variables.
// Keeping it dependency-free (no viper) makes the wiring obvious and easy to defend.
package config

import (
	"os"
	"strconv"
	"time"
)

// Common is configuration shared by every service.
type Common struct {
	MongoURI  string
	RedisAddr string
	// JWTSecret signs/verifies admin session tokens. Services that validate tokens (accounts, gateway)
	// must share the same value; it is fine to leave the dev default for the others.
	JWTSecret string
}

func LoadCommon() Common {
	return Common{
		MongoURI:  env("MONGO_URI", "mongodb://localhost:27017"),
		RedisAddr: env("REDIS_ADDR", "localhost:6379"),
		JWTSecret: env("JWT_SECRET", "dev-insecure-secret-change-me"),
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
