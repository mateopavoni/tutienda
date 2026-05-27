package orders

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// tenantHeaders forwards the active store id from the context so the downstream service stays scoped
// to the same store as the checkout. Empty when no tenant is set (the downstream then rejects it).
func tenantHeaders(ctx context.Context) map[string]string {
	if id, ok := tenant.FromContext(ctx); ok {
		return map[string]string{tenant.Header: id}
	}
	return nil
}

// OutOfStockError reports a SKU that could not be reserved by the inventory service.
type OutOfStockError struct{ SKU string }

func (e *OutOfStockError) Error() string { return "out of stock: " + e.SKU }

// ErrVariantNotFound means the catalog has no such SKU.
var ErrVariantNotFound = errors.New("variant not found")

// DropNotReleasedError reports an attempt to buy a SKU whose scheduled drop has not opened yet. It
// carries the release time so the storefront can show an accurate countdown on a rejected checkout.
type DropNotReleasedError struct {
	SKU    string
	DropAt time.Time
}

func (e *DropNotReleasedError) Error() string { return "drop not released: " + e.SKU }

// reserveLine is the wire shape inventory expects.
type reserveLine struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

// InventoryClient talks to the inventory service over HTTP.
type InventoryClient struct {
	c *httpx.Client
}

func NewInventoryClient(baseURL string) *InventoryClient {
	return &InventoryClient{c: httpx.NewClient(baseURL)}
}

// Reserve holds stock and returns the reservation id. A 409 is surfaced as OutOfStockError so the
// saga can fail fast without persisting an order.
func (ic *InventoryClient) Reserve(ctx context.Context, lines []reserveLine) (string, error) {
	body := map[string]any{"lines": lines}
	var res struct {
		ID string `json:"id"`
	}
	status, err := ic.c.DoWithHeaders(ctx, http.MethodPost, "/reservations", tenantHeaders(ctx), body, &res)
	if err != nil {
		var env *httpx.Error
		if status == http.StatusConflict && errors.As(err, &env) {
			return "", &OutOfStockError{SKU: env.SKU}
		}
		return "", err
	}
	return res.ID, nil
}

func (ic *InventoryClient) Confirm(ctx context.Context, reservationID string) error {
	_, err := ic.c.DoWithHeaders(ctx, http.MethodPost, "/reservations/"+reservationID+"/confirm", tenantHeaders(ctx), nil, nil)
	return err
}

func (ic *InventoryClient) Release(ctx context.Context, reservationID string) error {
	_, err := ic.c.DoWithHeaders(ctx, http.MethodPost, "/reservations/"+reservationID+"/release", tenantHeaders(ctx), nil, nil)
	return err
}

// catalogVariant mirrors catalog.VariantInfo on the wire.
type catalogVariant struct {
	SKU         string     `json:"sku"`
	ProductID   string     `json:"productId"`
	ProductName string     `json:"productName"`
	Label       string     `json:"label"`
	PriceCents  int64      `json:"priceCents"`
	Currency    string     `json:"currency"`
	DropAt      *time.Time `json:"dropAt,omitempty"`
}

// CatalogClient talks to the catalog service for server-authoritative pricing.
type CatalogClient struct {
	c *httpx.Client
}

func NewCatalogClient(baseURL string) *CatalogClient {
	return &CatalogClient{c: httpx.NewClient(baseURL)}
}

func (cc *CatalogClient) Variant(ctx context.Context, sku string) (*catalogVariant, error) {
	var v catalogVariant
	status, err := cc.c.DoWithHeaders(ctx, http.MethodGet, "/variants/"+sku, tenantHeaders(ctx), nil, &v)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, ErrVariantNotFound
		}
		return nil, err
	}
	return &v, nil
}
