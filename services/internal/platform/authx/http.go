package authx

import (
	"context"
	"net/http"
)

// MerchantHeader carries the authenticated merchant id from the gateway (which validates the JWT) to
// the accounts service. Like X-Tenant-ID, services trust it because only the gateway is public.
const MerchantHeader = "X-Merchant-ID"

type ctxKey struct{}

// WithMerchant returns a child context carrying the merchant id.
func WithMerchant(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// MerchantFromContext returns the merchant id and whether one was present.
func MerchantFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok && id != ""
}

// MerchantMiddleware lifts the X-Merchant-ID header into the context when present.
func MerchantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := r.Header.Get(MerchantHeader); id != "" {
			r = r.WithContext(WithMerchant(r.Context(), id))
		}
		next.ServeHTTP(w, r)
	})
}
