package gateway

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// stripIdentityHeaders removes any client-supplied identity headers so a caller can never spoof a
// tenant, merchant or plan: only the gateway is trusted to set these, from a resolved slug or a
// verified JWT. Spoofing X-Store-Plan would otherwise be a free upgrade to any paid feature.
func stripIdentityHeaders(r *http.Request) {
	r.Header.Del(tenant.Header)
	r.Header.Del(authx.MerchantHeader)
	r.Header.Del(plan.Header)
}

// resolveTenant is the storefront middleware: it maps the request's store slug to a tenant id and
// stamps X-Tenant-ID for the downstream service. No slug or an unknown slug is a 4xx.
func resolveTenant(res *Resolver, log *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			slug := slugFromRequest(r)
			if slug == "" {
				httpx.Fail(w, http.StatusBadRequest, "missing store")
				return
			}
			id, err := res.Resolve(r.Context(), slug)
			if errors.Is(err, ErrStoreNotFound) {
				httpx.Fail(w, http.StatusNotFound, "unknown store")
				return
			}
			if err != nil {
				log.Warn("tenant resolve failed", "slug", slug, "error", err)
				httpx.Fail(w, http.StatusBadGateway, "could not resolve store")
				return
			}
			r.Header.Set(tenant.Header, id)
			next.ServeHTTP(w, r)
		})
	}
}

// requireStore is the admin middleware for store-scoped services (catalog/inventory writes): it
// verifies the Bearer token, requires a store claim, and stamps both X-Tenant-ID and X-Merchant-ID.
func requireStore(issuer *authx.Issuer) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			claims, err := verifyBearer(issuer, r)
			if err != nil {
				httpx.Fail(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if claims.StoreID == "" {
				httpx.Fail(w, http.StatusForbidden, "select a store first")
				return
			}
			r.Header.Set(tenant.Header, claims.StoreID)
			r.Header.Set(authx.MerchantHeader, claims.MerchantID)
			// The plan is only trusted alongside a store claim; normalize it so a stale/odd value can't
			// unlock a paid feature downstream.
			r.Header.Set(plan.Header, string(plan.Normalize(claims.Plan)))
			next.ServeHTTP(w, r)
		})
	}
}

// requirePlatformAdmin guards the cross-tenant /api/platform routes: a valid token whose Role claim is
// RoleAdmin. A regular merchant token (or none) is rejected, so only the SaaS super-admin gets through.
func requirePlatformAdmin(issuer *authx.Issuer) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			claims, err := verifyBearer(issuer, r)
			if err != nil {
				httpx.Fail(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if claims.Role != authx.RoleAdmin {
				httpx.Fail(w, http.StatusForbidden, "platform admin only")
				return
			}
			r.Header.Set(authx.MerchantHeader, claims.MerchantID)
			next.ServeHTTP(w, r)
		})
	}
}

// accountsAuth guards the accounts service: signup, login and the public by-slug lookup pass through;
// everything else requires a valid token and gets X-Merchant-ID stamped from its claims.
func accountsAuth(issuer *authx.Issuer) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			if isPublicAccountsPath(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := verifyBearer(issuer, r)
			if err != nil {
				httpx.Fail(w, http.StatusUnauthorized, "authentication required")
				return
			}
			r.Header.Set(authx.MerchantHeader, claims.MerchantID)
			next.ServeHTTP(w, r)
		})
	}
}

// isPublicAccountsPath reports whether an accounts route needs no authentication. Paths still carry
// the public /api/accounts prefix here (the proxy strips it later).
func isPublicAccountsPath(method, path string) bool {
	switch {
	case strings.HasSuffix(path, "/signup"), strings.HasSuffix(path, "/login"):
		return true
	case method == http.MethodGet && strings.Contains(path, "/stores/by-slug/"):
		return true
	default:
		return false
	}
}

// verifyBearer extracts and validates the Bearer token from the Authorization header.
func verifyBearer(issuer *authx.Issuer, r *http.Request) (*authx.Claims, error) {
	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok || token == "" {
		return nil, errors.New("missing bearer token")
	}
	return issuer.Verify(strings.TrimSpace(token))
}
