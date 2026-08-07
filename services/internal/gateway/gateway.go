package gateway

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
)

// Config wires the gateway: where each service lives, the slug→tenant resolver, the JWT issuer used to
// verify admin tokens, the rate limiter and a logger.
type Config struct {
	CatalogURL   string
	InventoryURL string
	OrdersURL    string
	AccountsURL  string
	Resolver     *Resolver
	Issuer       *authx.Issuer
	Limiter      *RateLimiter
	Log          *slog.Logger
}

// Router builds the public gateway. Three families of routes, each with its own identity handling:
//
//   - Storefront (read + checkout): "/api/{catalog,inventory,orders}". The store is resolved from the
//     request slug and stamped as X-Tenant-ID. Public, rate-limited, tenant-scoped. catalog/inventory are
//     additionally locked to GET/HEAD here (readOnly) — the backends only check tenant.Require, so
//     writes must go through the store-JWT-gated admin mount instead.
//   - Admin (writes): "/api/admin/{catalog,inventory}". A store-scoped JWT is required; the tenant
//     comes from the signed token, so a merchant can only touch a store they own.
//   - Accounts (identity): "/api/accounts". signup/login/by-slug are public; the rest need a token.
//
// Both catalog/inventory mounts (public and admin) also block any request path containing
// "/admin/tenant" (blockInternalAdminPath): that route cascade-deletes a whole store's data and is meant
// to only be reachable on the internal docker network, never through this gateway.
//
// catalog and inventory expose resource paths under their own prefix (/products, /stock), so that
// prefix is stripped entirely; orders' own resource is already "/orders", so only "/api" is stripped.
// Each route is registered both bare and as a subtree to avoid the 301 redirect ServeMux issues for a
// bare path against a trailing-slash pattern.
func Router(cfg Config) (http.Handler, error) {
	catalog, err := newProxy(cfg.CatalogURL, "/api/catalog")
	if err != nil {
		return nil, err
	}
	inventory, err := newProxy(cfg.InventoryURL, "/api/inventory")
	if err != nil {
		return nil, err
	}
	ordersProxy, err := newProxy(cfg.OrdersURL, "/api")
	if err != nil {
		return nil, err
	}
	adminCatalog, err := newProxy(cfg.CatalogURL, "/api/admin/catalog")
	if err != nil {
		return nil, err
	}
	adminInventory, err := newProxy(cfg.InventoryURL, "/api/admin/inventory")
	if err != nil {
		return nil, err
	}
	accounts, err := newProxy(cfg.AccountsURL, "/api/accounts")
	if err != nil {
		return nil, err
	}
	// Platform admin and storefront customers both hit the accounts service; strip only "/api" so it
	// sees "/platform/..." / "/customers/..." (its own route prefixes).
	platform, err := newProxy(cfg.AccountsURL, "/api")
	if err != nil {
		return nil, err
	}
	customers, err := newProxy(cfg.AccountsURL, "/api")
	if err != nil {
		return nil, err
	}

	limit := rateLimit(cfg.Limiter, cfg.Log)
	storefront := resolveTenant(cfg.Resolver, cfg.Log)
	storeOrders := storefrontBuyer(cfg.Resolver, cfg.Issuer, cfg.Log)
	admin := requireStore(cfg.Issuer)
	identity := accountsAuth(cfg.Issuer)
	platformAdmin := requirePlatformAdmin(cfg.Issuer)
	buyer := customerAuth(cfg.Issuer, cfg.Resolver, cfg.Log)
	// noInternalAdmin blocks the catalog/inventory "/admin/tenant" cascade-delete route: it must only be
	// reachable on the internal docker network (accounts calls it directly), never through either the
	// public or the store-scoped gateway mount. See blockInternalAdminPath in auth.go.
	noInternalAdmin := blockInternalAdminPath()
	// readOnly restricts the public (unauthenticated, slug-resolved) catalog/inventory mount to safe read
	// methods — writes must go through /api/admin/{catalog,inventory}, which requires requireStore. See
	// readOnly in auth.go for why this matters: the backends only check tenant.Require, which trusts
	// whatever X-Tenant-ID the gateway stamped, so without this gate a bare X-Tenant-Slug would be enough
	// to write to any tenant.
	readOnlyMethods := readOnly()

	mux := http.NewServeMux()
	mount := func(prefix string, h http.Handler, mw ...httpx.Middleware) {
		wrapped := httpx.Chain(h, append([]httpx.Middleware{limit}, mw...)...)
		mux.Handle(prefix, wrapped)
		mux.Handle(prefix+"/", wrapped)
	}
	// Admin routes are registered before the storefront ones, but ServeMux matches the longest pattern
	// regardless of order; the distinct "/api/admin/..." prefix keeps them unambiguous.
	mount("/api/admin/catalog", adminCatalog, noInternalAdmin, admin)
	mount("/api/admin/inventory", adminInventory, noInternalAdmin, admin)
	mount("/api/catalog", catalog, noInternalAdmin, readOnlyMethods, storefront)
	mount("/api/inventory", inventory, noInternalAdmin, readOnlyMethods, storefront)
	mount("/api/orders", ordersProxy, storeOrders)
	mount("/api/accounts", accounts, identity)
	mount("/api/platform", platform, platformAdmin)
	mount("/api/customers", customers, buyer)
	mux.HandleFunc("GET /health", healthHandler(cfg))
	return mux, nil
}

// newProxy reverse-proxies to target, stripping the public prefix so the service sees its own paths
// (e.g. /api/catalog/products -> /products).
func newProxy(target, stripPrefix string) (http.Handler, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Director = func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.Host = u.Host
		p := strings.TrimPrefix(r.URL.Path, stripPrefix)
		if p == "" || p[0] != '/' {
			p = "/" + p
		}
		r.URL.Path = p
	}
	return proxy, nil
}

// rateLimit denies requests over the per-client window limit with a 429 + Retry-After.
func rateLimit(rl *RateLimiter, log *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := rl.Allow(r.Context(), clientIP(r))
			if err != nil {
				log.Warn("rate limiter degraded, allowing request", "error", err)
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.Limit()))
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(rl.Window().Seconds())))
				httpx.Fail(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP prefers the first X-Forwarded-For hop (set by an upstream proxy like Nginx) and falls
// back to the socket address.
func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.IndexByte(fwd, ','); i >= 0 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// healthHandler fans out to each service's /healthz and reports per-service status.
func healthHandler(cfg Config) http.HandlerFunc {
	targets := map[string]string{
		"catalog":   cfg.CatalogURL,
		"inventory": cfg.InventoryURL,
		"orders":    cfg.OrdersURL,
		"accounts":  cfg.AccountsURL,
	}
	client := &http.Client{Timeout: 2 * time.Second}

	return func(w http.ResponseWriter, r *http.Request) {
		services := make(map[string]string, len(targets))
		allOK := true
		for name, base := range targets {
			if probe(r.Context(), client, base+"/healthz") {
				services[name] = "ok"
			} else {
				services[name] = "unreachable"
				allOK = false
			}
		}
		status := http.StatusOK
		overall := "ok"
		if !allOK {
			status = http.StatusServiceUnavailable
			overall = "degraded"
		}
		httpx.JSON(w, status, map[string]any{"status": overall, "services": services})
	}
}

func probe(ctx context.Context, client *http.Client, url string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}
