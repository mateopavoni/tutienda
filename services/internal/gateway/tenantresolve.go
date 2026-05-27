package gateway

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
)

// ErrStoreNotFound means no store matches the requested slug.
var ErrStoreNotFound = errors.New("store not found")

// Resolver turns a public store slug into its tenant id by asking the accounts service, with a short
// Redis cache so the storefront's hot path doesn't hit accounts on every request.
type Resolver struct {
	accounts *httpx.Client
	cache    *redis.Client
	ttl      time.Duration
}

func NewResolver(accountsURL string, cache *redis.Client, ttl time.Duration) *Resolver {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Resolver{accounts: httpx.NewClient(accountsURL), cache: cache, ttl: ttl}
}

// Resolve returns the tenant id for a slug.
func (r *Resolver) Resolve(ctx context.Context, slug string) (string, error) {
	key := "tenant:slug:" + slug
	if r.cache != nil {
		if id, err := r.cache.Get(ctx, key).Result(); err == nil && id != "" {
			return id, nil
		}
	}
	var store struct {
		ID string `json:"id"`
	}
	status, err := r.accounts.Do(ctx, http.MethodGet, "/stores/by-slug/"+slug, nil, &store)
	if err != nil {
		if status == http.StatusNotFound {
			return "", ErrStoreNotFound
		}
		return "", err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, key, store.ID, r.ttl).Err()
	}
	return store.ID, nil
}

// slugFromRequest reads the store slug from the explicit X-Tenant-Slug header (used in dev and by the
// SvelteKit server) or, failing that, from the leftmost label of the Host (subdomain routing in prod).
func slugFromRequest(r *http.Request) string {
	if s := strings.TrimSpace(r.Header.Get("X-Tenant-Slug")); s != "" {
		return s
	}
	return subdomainSlug(r.Host)
}

// subdomainSlug returns the leftmost host label when the host looks like slug.domain.tld, and "" for
// bare hosts (localhost, an IP, or an apex domain) where there is no store subdomain.
func subdomainSlug(host string) string {
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i] // drop port
	}
	if host == "" || host == "localhost" {
		return ""
	}
	labels := strings.Split(host, ".")
	if len(labels) < 3 {
		return "" // apex domain or localhost-style host: no subdomain
	}
	switch labels[0] {
	case "www", "api", "admin":
		return ""
	default:
		return labels[0]
	}
}
