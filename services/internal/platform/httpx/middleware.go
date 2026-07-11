package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

// Middleware is the standard decorator shape for net/http handlers.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order, so Chain(h, a, b) runs a, then b, then h.
func Chain(h http.Handler, mw ...Middleware) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// statusRecorder captures the response status code for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logger emits one structured access log line per request.
func Logger(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// Recover turns a panic into a 500 instead of crashing the server.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "error", rec, "path", r.URL.Path)
					Fail(w, http.StatusInternalServerError, "internal error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders sets the baseline response headers OWASP recommends for an API that is reachable
// directly from a browser (storefront fetch, admin dashboard). The gateway is the single public entry
// point, so this is the one place these need to be set for every downstream service's response.
//   - X-Content-Type-Options: stops the browser from MIME-sniffing a JSON error body into something
//     executable.
//   - X-Frame-Options / frame-ancestors: the API serves no HTML, but a same-origin-policy-confused
//     browser plugin or a misconfigured proxy embedding it in a frame is cheap to close off anyway.
//   - Referrer-Policy: never leak the full request URL (which can carry a slug/id) to a third party.
//   - Permissions-Policy: the API needs none of these browser features; deny them explicitly.
//   - Strict-Transport-Security: only meaningful once the deploy terminates TLS (Nginx/Dokku in front),
//     so it is a no-op over plain HTTP in local dev but correct in production.
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			next.ServeHTTP(w, r)
		})
	}
}

// CORS allows the storefront origin and the methods/headers the API uses. The admin panel adds write
// verbs (PUT/PATCH/DELETE), a Bearer Authorization header and the X-Tenant-Slug store selector.
func CORS(origin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-Slug")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
