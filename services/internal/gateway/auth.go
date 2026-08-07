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
	r.Header.Del(authx.CustomerHeader)
	r.Header.Del(plan.Header)
}

// resolveTenant is the storefront middleware: it maps the request's store slug to a tenant id and
// stamps X-Tenant-ID for the downstream service. No slug or an unknown slug is a 4xx.
func resolveTenant(res *Resolver, log *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			id, ok := resolveSlug(w, r, res, log)
			if !ok {
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
			if err != nil || claims.Aud == authx.AudCustomer {
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

// requireAdminRole writes 403 and returns false unless claims carry the platform super-admin role.
// Shared by requirePlatformAdmin (the whole /api/platform mount) and accountsAuth (the /api/accounts
// mount's /platform/* alias, see isAccountsPlatformPath) so the two gates can never diverge again.
func requireAdminRole(w http.ResponseWriter, claims *authx.Claims) bool {
	if claims.Role != authx.RoleAdmin {
		httpx.Fail(w, http.StatusForbidden, "platform admin only")
		return false
	}
	return true
}

// requirePlatformAdmin guards the cross-tenant /api/platform routes: a valid token whose Role claim is
// RoleAdmin. A regular merchant token (or none) is rejected, so only the SaaS super-admin gets through.
func requirePlatformAdmin(issuer *authx.Issuer) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			claims, err := verifyBearer(issuer, r)
			if err != nil || claims.Aud == authx.AudCustomer {
				httpx.Fail(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if !requireAdminRole(w, claims) {
				return
			}
			r.Header.Set(authx.MerchantHeader, claims.MerchantID)
			next.ServeHTTP(w, r)
		})
	}
}

// accountsAuth guards the accounts service: signup, login and the public by-slug lookup pass through;
// everything else requires a valid token and gets X-Merchant-ID stamped from its claims.
//
// The accounts service also registers the cross-tenant /platform/* routes (stats, all stores, all
// merchants, cross-tenant plan override) on the very same mux as /stores/*, and this mount reaches them
// too — it strips only "/api", same as the dedicated /api/platform mount does. Those handlers trust the
// gateway to have already checked Role == RoleAdmin (see requirePlatformAdmin) and don't re-check
// ownership themselves, so this mount must enforce the exact same rule for its /platform/* alias, or any
// merchant JWT holder reaches super-admin-only data through it.
func accountsAuth(issuer *authx.Issuer) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			if isPublicAccountsPath(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := verifyBearer(issuer, r)
			if err != nil || claims.Aud == authx.AudCustomer {
				httpx.Fail(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if isAccountsPlatformPath(r.URL.Path) && !requireAdminRole(w, claims) {
				return
			}
			r.Header.Set(authx.MerchantHeader, claims.MerchantID)
			next.ServeHTTP(w, r)
		})
	}
}

// customerAuth guards the storefront customer routes (/api/customers/*). signup/login are public but
// still need a tenant: the slug is resolved to X-Tenant-ID like the storefront. Everything else (me)
// requires a valid buyer token whose audience is AudCustomer; the tenant then comes from the token's own
// claim (not the slug), and the customer id is stamped as X-Customer-ID.
func customerAuth(issuer *authx.Issuer, res *Resolver, log *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			if strings.HasSuffix(r.URL.Path, "/signup") || strings.HasSuffix(r.URL.Path, "/login") {
				id, ok := resolveSlug(w, r, res, log)
				if !ok {
					return
				}
				r.Header.Set(tenant.Header, id)
				next.ServeHTTP(w, r)
				return
			}
			claims, err := verifyBearer(issuer, r)
			if err != nil || claims.Aud != authx.AudCustomer || claims.TenantID == "" {
				httpx.Fail(w, http.StatusUnauthorized, "authentication required")
				return
			}
			r.Header.Set(tenant.Header, claims.TenantID)
			r.Header.Set(authx.CustomerHeader, claims.MerchantID)
			next.ServeHTTP(w, r)
		})
	}
}

// storefrontBuyer is the storefront tenant resolver for the orders routes, plus an *optional* buyer
// stamp: if the caller carries a valid customer token bound to this same store, X-Customer-ID is set so
// the order is attributed to them. No token (or one for another store) ⇒ a guest checkout, still allowed.
func storefrontBuyer(res *Resolver, issuer *authx.Issuer, log *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stripIdentityHeaders(r)
			id, ok := resolveSlug(w, r, res, log)
			if !ok {
				return
			}
			r.Header.Set(tenant.Header, id)
			// Soft attribution: only honor a customer token that is for *this* store, so a buyer of store
			// A can never have an order in store B pinned to them.
			if claims, err := verifyBearer(issuer, r); err == nil && claims.Aud == authx.AudCustomer && claims.TenantID == id {
				r.Header.Set(authx.CustomerHeader, claims.MerchantID)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// resolveSlug maps the request's store slug to a tenant id, writing the proper 4xx/5xx and returning
// ok=false on failure. Shared by the storefront-family middlewares.
func resolveSlug(w http.ResponseWriter, r *http.Request, res *Resolver, log *slog.Logger) (string, bool) {
	slug := slugFromRequest(r)
	if slug == "" {
		httpx.Fail(w, http.StatusBadRequest, "missing store")
		return "", false
	}
	id, err := res.Resolve(r.Context(), slug)
	if errors.Is(err, ErrStoreNotFound) {
		httpx.Fail(w, http.StatusNotFound, "unknown store")
		return "", false
	}
	if errors.Is(err, ErrStoreDisabled) {
		httpx.Fail(w, http.StatusNotFound, "store unavailable")
		return "", false
	}
	if err != nil {
		log.Warn("tenant resolve failed", "slug", slug, "error", err)
		httpx.Fail(w, http.StatusBadGateway, "could not resolve store")
		return "", false
	}
	return id, true
}

// isPublicAccountsPath reports whether an accounts route needs no authentication. Paths still carry
// the public /api/accounts prefix here (the proxy strips it later).
func isPublicAccountsPath(method, path string) bool {
	switch {
	case strings.HasSuffix(path, "/signup"), strings.HasSuffix(path, "/login"):
		return true
	case method == http.MethodGet && strings.Contains(path, "/stores/by-slug/"):
		return true
	case method == http.MethodPost && strings.HasSuffix(path, "/messages"):
		return true
	default:
		return false
	}
}

// isAccountsPlatformPath reports whether a path reaching the /api/accounts mount targets the
// cross-tenant /platform/* routes that the accounts service also serves off the same mux (see
// accountsAuth's doc comment). Paths still carry the public /api/accounts prefix here, same as
// isPublicAccountsPath.
func isAccountsPlatformPath(path string) bool {
	rest := strings.TrimPrefix(path, "/api/accounts")
	return rest == "/platform" || strings.HasPrefix(rest, "/platform/")
}

// blockInternalAdminPath rejects any request whose path targets a backend's internal-only
// "/admin/tenant" route (catalog/inventory's cascade-delete-a-store endpoint, called by accounts over
// the internal docker network). It must never be reachable through the public gateway — not even with a
// valid store JWT — so this is checked ahead of any auth/tenant middleware on both the public and the
// admin catalog/inventory mounts. 404, not 403/405, so its existence isn't confirmed to a prober.
func blockInternalAdminPath() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/admin/tenant") {
				httpx.Fail(w, http.StatusNotFound, "not found")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// readOnly restricts a mount to safe read methods (GET/HEAD), rejecting everything else with 405. Used
// on the public /api/catalog and /api/inventory mounts: those backends only check tenant.Require (no
// auth), trusting whatever X-Tenant-ID the gateway stamped from a client-supplied, unauthenticated slug —
// so without this gate, any caller could write to any tenant's catalog/inventory by just naming its slug.
// Writes belong on /api/admin/{catalog,inventory}, which requires a store-scoped JWT (requireStore).
func readOnly() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				httpx.Fail(w, http.StatusMethodNotAllowed, "read-only")
				return
			}
			next.ServeHTTP(w, r)
		})
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
