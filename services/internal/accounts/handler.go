package accounts

import (
	"errors"
	"net/http"

	"github.com/mateopavoni/archive-commerce/internal/platform/authx"
	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// Handler exposes the accounts service over HTTP. Routes under /stores assume the gateway has already
// validated the session token and stamped X-Merchant-ID; /signup, /login and /stores/by-slug are public.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /signup", h.signup)
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("GET /stores", h.listStores)
	mux.HandleFunc("POST /stores", h.createStore)
	mux.HandleFunc("POST /stores/{id}/token", h.storeToken)
	mux.HandleFunc("PATCH /stores/{id}/plan", h.changePlan)
	mux.HandleFunc("PATCH /stores/{id}/status", h.changeStatus)
	mux.HandleFunc("PATCH /stores/{id}", h.updateStore)
	mux.HandleFunc("DELETE /stores/{id}", h.deleteStore)
	mux.HandleFunc("GET /stores/by-slug/{slug}", h.storeBySlug)
	// Storefront customers (buyers). Tenant comes from the gateway-resolved slug (signup/login) or the
	// buyer token's tenant claim (me); never from the client. Distinct buyer-token audience, enforced at
	// the gateway, keeps these from crossing with merchant routes.
	mux.HandleFunc("POST /customers/signup", h.customerSignup)
	mux.HandleFunc("POST /customers/login", h.customerLogin)
	mux.HandleFunc("GET /customers/me", h.customerMe)
	// Platform super-admin (cross-tenant). The gateway only routes these after verifying RoleAdmin, so
	// here they simply trust X-Merchant-ID like every other authenticated route.
	mux.HandleFunc("GET /platform/stats", h.platformStats)
	mux.HandleFunc("GET /platform/stores", h.platformStores)
	mux.HandleFunc("GET /platform/merchants", h.platformMerchants)
	mux.HandleFunc("PATCH /platform/stores/{id}/plan", h.platformSetPlan)
	// Lift every gateway-stamped identity header into the context; each handler reads only the one it
	// needs (merchant for /stores+/platform, tenant+customer for /customers). Lifting all is harmless.
	return tenant.Middleware()(authx.MerchantMiddleware(authx.CustomerMiddleware(mux)))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "accounts"})
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type session struct {
	Token    string    `json:"token"`
	Merchant *Merchant `json:"merchant,omitempty"`
	Store    *Store    `json:"store,omitempty"`
}

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := httpx.Decode(r, &c); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, m, err := h.svc.Signup(r.Context(), c.Email, c.Password)
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpx.FailDetail(w, http.StatusBadRequest, "invalid input", err.Error())
		return
	case errors.Is(err, ErrDuplicate):
		httpx.Fail(w, http.StatusConflict, "email already registered")
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "signup failed")
		return
	}
	httpx.JSON(w, http.StatusCreated, session{Token: token, Merchant: m})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := httpx.Decode(r, &c); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, m, err := h.svc.Login(r.Context(), c.Email, c.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.Fail(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "login failed")
		return
	}
	httpx.JSON(w, http.StatusOK, session{Token: token, Merchant: m})
}

func (h *Handler) listStores(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	stores, err := h.svc.MyStores(r.Context(), ownerID)
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not list stores")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": stores})
}

func (h *Handler) createStore(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	var body struct {
		Slug        string `json:"slug"`
		DisplayName string `json:"displayName"`
		Theme       string `json:"theme"`
		AccentColor string `json:"accentColor"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	store, err := h.svc.CreateStore(r.Context(), ownerID, body.Slug, body.DisplayName, body.Theme, body.AccentColor)
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpx.FailDetail(w, http.StatusBadRequest, "invalid input", err.Error())
		return
	case errors.Is(err, ErrDuplicate):
		httpx.Fail(w, http.StatusConflict, "slug already taken")
		return
	case errors.Is(err, ErrStoreLimit):
		httpx.FailDetail(w, http.StatusConflict, "store limit reached", err.Error())
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "could not create store")
		return
	}
	httpx.JSON(w, http.StatusCreated, store)
}

func (h *Handler) storeToken(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	token, store, err := h.svc.StoreToken(r.Context(), ownerID, r.PathValue("id"))
	if errors.Is(err, ErrStoreNotFound) {
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	httpx.JSON(w, http.StatusOK, session{Token: token, Store: store})
}

func (h *Handler) updateStore(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	var body struct {
		DisplayName string   `json:"displayName"`
		Settings    Settings `json:"settings"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	store, err := h.svc.UpdateStore(r.Context(), ownerID, r.PathValue("id"), body.DisplayName, body.Settings)
	if errors.Is(err, ErrStoreNotFound) {
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not update store")
		return
	}
	httpx.JSON(w, http.StatusOK, store)
}

func (h *Handler) deleteStore(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	err := h.svc.DeleteStore(r.Context(), ownerID, r.PathValue("id"))
	if errors.Is(err, ErrStoreNotFound) {
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not delete store")
		return
	}
	httpx.JSON(w, http.StatusNoContent, nil)
}

func (h *Handler) changePlan(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	var body struct {
		Plan string `json:"plan"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	store, err := h.svc.SetPlan(r.Context(), ownerID, r.PathValue("id"), body.Plan)
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpx.FailDetail(w, http.StatusBadRequest, "invalid plan", err.Error())
		return
	case errors.Is(err, ErrStoreNotFound):
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "could not change plan")
		return
	}
	httpx.JSON(w, http.StatusOK, store)
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := h.merchant(w, r)
	if !ok {
		return
	}
	var body struct {
		Disabled bool `json:"disabled"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	store, err := h.svc.SetDisabled(r.Context(), ownerID, r.PathValue("id"), body.Disabled)
	if errors.Is(err, ErrStoreNotFound) {
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not change store status")
		return
	}
	httpx.JSON(w, http.StatusOK, store)
}

func (h *Handler) storeBySlug(w http.ResponseWriter, r *http.Request) {
	store, err := h.svc.StoreBySlug(r.Context(), r.PathValue("slug"))
	if errors.Is(err, ErrStoreNotFound) {
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not resolve store")
		return
	}
	httpx.JSON(w, http.StatusOK, store)
}

func (h *Handler) platformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not load stats")
		return
	}
	httpx.JSON(w, http.StatusOK, stats)
}

func (h *Handler) platformStores(w http.ResponseWriter, r *http.Request) {
	stores, err := h.svc.AllStores(r.Context())
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not list stores")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": stores})
}

func (h *Handler) platformMerchants(w http.ResponseWriter, r *http.Request) {
	merchants, err := h.svc.AllMerchants(r.Context())
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not list merchants")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": merchants})
}

func (h *Handler) platformSetPlan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Plan string `json:"plan"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	store, err := h.svc.SetPlanAsAdmin(r.Context(), r.PathValue("id"), body.Plan)
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpx.FailDetail(w, http.StatusBadRequest, "invalid plan", err.Error())
		return
	case errors.Is(err, ErrStoreNotFound):
		httpx.Fail(w, http.StatusNotFound, "store not found")
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "could not change plan")
		return
	}
	httpx.JSON(w, http.StatusOK, store)
}

// merchant pulls the gateway-injected merchant id, writing 401 when absent.
func (h *Handler) merchant(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := authx.MerchantFromContext(r.Context())
	if !ok {
		httpx.Fail(w, http.StatusUnauthorized, "authentication required")
		return "", false
	}
	return id, true
}

// --- storefront customers (buyers) ---

type customerSession struct {
	Token    string    `json:"token"`
	Customer *Customer `json:"customer,omitempty"`
}

func (h *Handler) customerSignup(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, c, err := h.svc.CustomerSignup(r.Context(), tid, body.Email, body.Password, body.Name)
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpx.FailDetail(w, http.StatusBadRequest, "invalid input", err.Error())
		return
	case errors.Is(err, ErrDuplicate):
		httpx.Fail(w, http.StatusConflict, "email already registered")
		return
	case err != nil:
		httpx.Fail(w, http.StatusInternalServerError, "signup failed")
		return
	}
	httpx.JSON(w, http.StatusCreated, customerSession{Token: token, Customer: c})
}

func (h *Handler) customerLogin(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, c, err := h.svc.CustomerLogin(r.Context(), tid, body.Email, body.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		httpx.Fail(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "login failed")
		return
	}
	httpx.JSON(w, http.StatusOK, customerSession{Token: token, Customer: c})
}

func (h *Handler) customerMe(w http.ResponseWriter, r *http.Request) {
	tid, ok := tenant.Require(w, r)
	if !ok {
		return
	}
	cid, ok := authx.CustomerFromContext(r.Context())
	if !ok {
		httpx.Fail(w, http.StatusUnauthorized, "authentication required")
		return
	}
	c, err := h.svc.Customer(r.Context(), tid, cid)
	if errors.Is(err, ErrCustomerNotFound) {
		httpx.Fail(w, http.StatusNotFound, "customer not found")
		return
	}
	if err != nil {
		httpx.Fail(w, http.StatusInternalServerError, "could not read customer")
		return
	}
	httpx.JSON(w, http.StatusOK, c)
}
