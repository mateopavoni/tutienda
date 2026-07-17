package catalog

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/media"
	"github.com/mateopavoni/archive-commerce/internal/platform/plan"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// Handler exposes the catalog service over HTTP.
type Handler struct {
	svc   *Service
	media *media.Client // may be nil when no object store is configured; uploads then return 503.
}

func NewHandler(svc *Service, media *media.Client) *Handler {
	return &Handler{svc: svc, media: media}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /products", h.list)
	mux.HandleFunc("GET /products/{id}", h.get)
	mux.HandleFunc("GET /variants/{sku}", h.variant)
	mux.HandleFunc("POST /products", h.create)
	mux.HandleFunc("PUT /products/{id}", h.update)
	mux.HandleFunc("DELETE /products/{id}", h.remove)
	mux.HandleFunc("POST /uploads", h.upload)
	mux.HandleFunc("POST /seed", h.seed)
	mux.HandleFunc("DELETE /admin/tenant", h.removeTenant)
	// Lift the gateway-stamped X-Tenant-ID and X-Store-Plan into the context for every handler. The plan
	// only rides admin (store-scoped) requests; storefront reads carry none and read back as free, which
	// is fine since reads aren't gated.
	return httpx.Chain(mux, tenant.Middleware(), plan.Middleware())
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "catalog"})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	category := r.URL.Query().Get("category")
	limit := int64(50)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	products, err := h.svc.List(r.Context(), tid, category, limit)
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not list products")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": products})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	p, err := h.svc.Get(r.Context(), tid, r.PathValue("id"))
	if errors.Is(err, ErrProductNotFound) {
		httpx.Fail(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not read product")
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) variant(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	v, err := h.svc.Variant(r.Context(), tid, r.PathValue("sku"))
	if errors.Is(err, ErrProductNotFound) {
		httpx.Fail(w, http.StatusNotFound, "variant not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not read variant")
		return
	}
	httpx.JSON(w, http.StatusOK, v)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	var p Product
	if err := httpx.Decode(r, &p); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid product body")
		return
	}
	created, err := h.svc.Create(r.Context(), tid, p)
	switch {
	case errors.Is(err, ErrInvalidProduct):
		httpx.FailDetail(w, http.StatusBadRequest, "invalid product", err.Error())
		return
	case errors.Is(err, ErrFeatureLocked), errors.Is(err, ErrProductLimit):
		httpx.FailDetail(w, http.StatusPaymentRequired, "upgrade required", err.Error())
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "could not create product")
		return
	}
	httpx.JSON(w, http.StatusCreated, created)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	var p Product
	if err := httpx.Decode(r, &p); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid product body")
		return
	}
	updated, err := h.svc.Update(r.Context(), tid, r.PathValue("id"), p)
	switch {
	case errors.Is(err, ErrProductNotFound):
		httpx.Fail(w, http.StatusNotFound, "product not found")
		return
	case errors.Is(err, ErrFeatureLocked):
		httpx.FailDetail(w, http.StatusPaymentRequired, "upgrade required", err.Error())
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "could not update product")
		return
	}
	httpx.JSON(w, http.StatusOK, updated)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	err := h.svc.Delete(r.Context(), tid, r.PathValue("id"))
	if errors.Is(err, ErrProductNotFound) {
		httpx.Fail(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not delete product")
		return
	}
	httpx.JSON(w, http.StatusNoContent, nil)
}

// removeTenant deletes every product of a store, an internal call from accounts when a store itself is
// deleted (cascade). Not reachable through the gateway — only over the service's own internal network.
func (h *Handler) removeTenant(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteTenant(r.Context(), tid); err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not delete store products")
		return
	}
	httpx.JSON(w, http.StatusNoContent, nil)
}

// maxUploadBytes caps an image upload (10 MiB) so a merchant can't push an unbounded blob through the
// storefront builder — high enough for a full-quality hero/gallery photo. allowedImageExt maps accepted
// content types to a file extension — the closed set keeps non-image uploads (and surprising content
// types) out of the public bucket.
const maxUploadBytes = 10 << 20

var allowedImageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/avif": ".avif",
	"image/gif":  ".gif",
}

// upload stores a product/section image in the object store and returns its public URL. Tenant-scoped:
// the object lands under the caller's store prefix, so stores never collide. Plan gating is not applied
// here — images are table stakes for any store; the paid gates live on drops and the product cap.
func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	if h.media == nil {
		httpx.Fail(w, http.StatusServiceUnavailable, "image storage is not configured")
		return
	}
	// Cap the body before parsing so an oversized upload is rejected without buffering it all.
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Fail(w, http.StatusBadRequest, "expected a multipart 'file' field (max 10MB)")
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	ext, allowed := allowedImageExt[contentType]
	if !allowed {
		httpx.Fail(w, http.StatusUnsupportedMediaType, "only JPEG, PNG, WebP, AVIF or GIF images are accepted")
		return
	}
	url, err := h.media.Upload(r.Context(), tid, ext, contentType, file, header.Size)
	if err != nil {
		httpx.Fail(w, http.StatusBadGateway, "could not store the image")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"url": url})
}

func (h *Handler) seed(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.Seed(r.Context())
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "seed failed")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"seeded": n})
}
