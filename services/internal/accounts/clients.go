package accounts

import (
	"context"
	"net/http"

	"github.com/mateopavoni/archive-commerce/internal/platform/httpx"
	"github.com/mateopavoni/archive-commerce/internal/platform/tenant"
)

// tenantHeaders stamps the store id being deleted so the downstream service scopes its cascade delete to
// that tenant only — same pattern as orders' clients.go.
func tenantHeaders(storeID string) map[string]string {
	return map[string]string{tenant.Header: storeID}
}

// CatalogClient talks to the catalog service to cascade-delete a store's products.
type CatalogClient struct{ c *httpx.Client }

func NewCatalogClient(baseURL string) *CatalogClient {
	return &CatalogClient{c: httpx.NewClient(baseURL)}
}

func (cc *CatalogClient) DeleteTenant(ctx context.Context, storeID string) error {
	_, err := cc.c.DoWithHeaders(ctx, http.MethodDelete, "/admin/tenant", tenantHeaders(storeID), nil, nil)
	return err
}

// InventoryClient talks to the inventory service to cascade-delete a store's stock and reservations.
type InventoryClient struct{ c *httpx.Client }

func NewInventoryClient(baseURL string) *InventoryClient {
	return &InventoryClient{c: httpx.NewClient(baseURL)}
}

func (ic *InventoryClient) DeleteTenant(ctx context.Context, storeID string) error {
	_, err := ic.c.DoWithHeaders(ctx, http.MethodDelete, "/admin/tenant", tenantHeaders(storeID), nil, nil)
	return err
}
