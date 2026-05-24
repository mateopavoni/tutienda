// Package tenant carries the active store (tenant) through a request: the gateway resolves it and
// stamps the X-Tenant-ID header, the middleware lifts it into the context, and every service reads it
// from there. This is the single thread that scopes shared collections to one store.
package tenant

import (
	"context"
	"net/http"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
)

// Header is the wire name the gateway uses to pass the resolved tenant id downstream.
const Header = "X-Tenant-ID"

type ctxKey struct{}

// WithTenant returns a child context carrying the tenant id.
func WithTenant(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the tenant id and whether one was present.
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok && id != ""
}

// Middleware lifts the X-Tenant-ID header into the request context when present. It does not reject
// missing tenants — read-only public endpoints (e.g. /healthz) have none — so handlers that require a
// tenant must call Require.
func Middleware() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if id := r.Header.Get(Header); id != "" {
				r = r.WithContext(WithTenant(r.Context(), id))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Require pulls the tenant id from the context, writing a 400 and returning ok=false when absent. A
// missing tenant on a tenant-scoped route means the gateway failed to resolve a store — a client error.
func Require(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := FromContext(r.Context())
	if !ok {
		httpx.Fail(w, http.StatusBadRequest, "missing tenant")
		return "", false
	}
	return id, true
}
