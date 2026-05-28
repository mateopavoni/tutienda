import { env } from '$env/dynamic/private';
import type { Product, Stock } from '$lib/types';
import type { StoreInfo } from '$lib/tenant';

// Server-side client used by load functions during SSR. It talks to the gateway over the internal
// network (INTERNAL_API_BASE), which is reachable from the web container but not from the browser.
// Every storefront call carries X-Tenant-Slug so the gateway scopes it to the right store.
function base(): string {
	return env.INTERNAL_API_BASE || 'http://localhost:8080';
}

type Fetch = typeof fetch;

function tenantHeaders(slug: string): HeadersInit {
	return { 'X-Tenant-Slug': slug };
}

export async function getStore(fetchFn: Fetch, slug: string): Promise<StoreInfo | null> {
	const res = await fetchFn(base() + '/api/accounts/stores/by-slug/' + encodeURIComponent(slug));
	if (res.status === 404) return null;
	if (!res.ok) throw new Error('store resolve failed: ' + res.status);
	return (await res.json()) as StoreInfo;
}

export async function listProducts(fetchFn: Fetch, slug: string, category?: string): Promise<Product[]> {
	const query = category ? '?category=' + encodeURIComponent(category) : '';
	const res = await fetchFn(base() + '/api/catalog/products' + query, { headers: tenantHeaders(slug) });
	if (!res.ok) throw new Error('catalog list failed: ' + res.status);
	const data = (await res.json()) as { items: Product[] };
	return data.items ?? [];
}

export async function getProduct(fetchFn: Fetch, slug: string, id: string): Promise<Product | null> {
	const res = await fetchFn(base() + '/api/catalog/products/' + encodeURIComponent(id), {
		headers: tenantHeaders(slug)
	});
	if (res.status === 404) return null;
	if (!res.ok) throw new Error('catalog get failed: ' + res.status);
	return (await res.json()) as Product;
}

export async function getStock(fetchFn: Fetch, slug: string, skus: string[]): Promise<Stock[]> {
	if (skus.length === 0) return [];
	const res = await fetchFn(base() + '/api/inventory/stock?skus=' + skus.join(','), {
		headers: tenantHeaders(slug)
	});
	if (!res.ok) throw new Error('stock read failed: ' + res.status);
	const data = (await res.json()) as { items: Stock[] };
	return data.items ?? [];
}
