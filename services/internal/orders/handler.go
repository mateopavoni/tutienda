package orders

import (
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// Handler exposes the orders service over HTTP.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /orders", h.checkout)
	mux.HandleFunc("GET /orders/mine", h.mine)
	mux.HandleFunc("GET /orders/{id}", h.get)
	// Lift the gateway-stamped X-Tenant-ID and (optional) X-Customer-ID into the context for every handler.
	return tenant.Middleware()(authx.CustomerMiddleware(mux))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "orders"})
}

type checkoutRequest struct {
	Items    []CheckoutItem  `json:"items"`
	Shipping ShippingAddress `json:"shipping"`
	// Approve simulates the payment gateway's decision (portfolio project, no real gateway).
	Approve bool `json:"approve"`
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.svc.Checkout(r.Context(), req.Items, req.Shipping, req.Approve)
	if err != nil {
		var oos *OutOfStockError
		var drop *DropNotReleasedError
		switch {
		case errors.As(err, &oos):
			httpx.JSON(w, http.StatusConflict, httpx.Error{Message: "out of stock", SKU: oos.SKU})
		case errors.As(err, &drop):
			httpx.JSON(w, http.StatusConflict, httpx.Error{Message: "drop not released", SKU: drop.SKU})
		case errors.Is(err, ErrInvalidShipping):
			httpx.Fail(w, http.StatusBadRequest, "invalid shipping address")
		case errors.Is(err, ErrVariantNotFound):
			httpx.FailDetail(w, http.StatusBadRequest, "unknown item", err.Error())
		default:
			httpx.FailDetail(w, http.StatusBadRequest, "checkout failed", err.Error())
		}
		return
	}
	httpx.JSON(w, http.StatusCreated, order)
}

// mine returns the authenticated customer's order history. Requires X-Customer-ID (the gateway stamps it
// only for a valid buyer token of this store), so a guest gets 401.
func (h *Handler) mine(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	cid, ok := authx.CustomerFromContext(r.Context())
	if !ok {
		httpx.Fail(w, http.StatusUnauthorized, "authentication required")
		return
	}
	orders, err := h.svc.Mine(r.Context(), tid, cid)
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not list orders")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": orders})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.svc.Get(r.Context(), tid, id)
	if errors.Is(err, ErrOrderNotFound) {
		httpx.Fail(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not read order")
		return
	}
	httpx.JSON(w, http.StatusOK, order)
}
