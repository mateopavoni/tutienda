package inventory

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// Handler exposes the inventory service over HTTP.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes returns the service mux. Go 1.22 pattern routing keeps this dependency-free.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /stock", h.getStock)
	mux.HandleFunc("POST /stock", h.setStock)
	mux.HandleFunc("POST /reservations", h.reserve)
	mux.HandleFunc("POST /reservations/{id}/confirm", h.confirm)
	mux.HandleFunc("POST /reservations/{id}/release", h.release)
	mux.HandleFunc("POST /seed", h.seed)
	// Lift the gateway-stamped X-Tenant-ID into the context for every handler.
	return tenant.Middleware()(mux)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "inventory"})
}

type reserveRequest struct {
	Lines      []Line `json:"lines"`
	TTLSeconds int    `json:"ttlSeconds,omitempty"`
}

func (h *Handler) reserve(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	var req reserveRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ttl := time.Duration(req.TTLSeconds) * time.Second
	res, err := h.svc.Reserve(r.Context(), tid, req.Lines, ttl)
	if err != nil {
		var oos *OutOfStockError
		if errors.As(err, &oos) {
			httpx.JSON(w, http.StatusConflict, httpx.Error{Message: "out of stock", SKU: oos.SKU})
			return
		}
		httpx.FailDetail(w, http.StatusBadRequest, "could not reserve", err.Error())
		return
	}
	httpx.JSON(w, http.StatusCreated, res)
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	h.finalize(w, r, h.svc.Confirm)
}

func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	h.finalize(w, r, h.svc.Release)
}

// finalize is the shared confirm/release flow: resolve the tenant, parse the id, run the action, map errors.
func (h *Handler) finalize(w http.ResponseWriter, r *http.Request, action func(context.Context, string, primitive.ObjectID) (*Reservation, error)) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid reservation id")
		return
	}
	res, err := action(r.Context(), tid, id)
	switch {
	case errors.Is(err, ErrReservationNotFound):
		httpx.Fail(w, http.StatusNotFound, "reservation not found")
	case errors.Is(err, ErrNotHeld):
		httpx.FailDetail(w, http.StatusConflict, "reservation not held", err.Error())
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "could not update reservation")
	default:
		httpx.JSON(w, http.StatusOK, res)
	}
}

func (h *Handler) getStock(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	raw := r.URL.Query().Get("skus")
	if raw == "" {
		httpx.Fail(w, http.StatusBadRequest, "skus query parameter is required")
		return
	}
	skus := strings.Split(raw, ",")
	stocks, err := h.svc.Stocks(r.Context(), tid, skus)
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not read stock")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": stocks})
}

// setStock is the admin endpoint to set a SKU's on-hand amount for the caller's store.
func (h *Handler) setStock(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	var body struct {
		SKU       string `json:"sku"`
		Available int    `json:"available"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SetStock(r.Context(), tid, body.SKU, body.Available); err != nil {
		httpx.FailDetail(w, http.StatusBadRequest, "could not set stock", err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"sku": body.SKU, "available": body.Available})
}

func (h *Handler) seed(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.SeedDefault(r.Context())
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "seed failed")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"seeded": n})
}
