import { env } from '$env/dynamic/public';
import type { ApiError, Order, Stock } from '$lib/types';
import { getBrowserSlug } from '$lib/tenant';
import { customerToken } from '$lib/stores/customer';

// Browser-side client. Mutations and live stock reads go through the public gateway, which applies
// CORS and rate limiting. SSR reads use the server client instead (src/lib/server/api.ts). Every call
// carries X-Tenant-Slug so the gateway scopes it to the active store.
function base(): string {
	return env.PUBLIC_API_BASE || 'http://localhost:8080';
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const headers = new Headers(init?.headers);
	headers.set('X-Tenant-Slug', getBrowserSlug());
	// If a buyer is signed in to this store, attach their token so the order is attributed to them. The
	// gateway only honors it when the token's tenant matches this store; otherwise it's a guest checkout.
	const token = customerToken();
	if (token) headers.set('Authorization', 'Bearer ' + token);
	const res = await fetch(base() + path, { ...init, headers });
	const data = (await res.json().catch(() => ({}))) as unknown;
	if (!res.ok) {
		throw data as ApiError;
	}
	return data as T;
}

export async function checkout(items: { sku: string; qty: number }[]): Promise<Order> {
	return request<Order>('/api/orders', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ items })
	});
}

export async function fetchStock(skus: string[]): Promise<Stock[]> {
	if (skus.length === 0) return [];
	const data = await request<{ items: Stock[] }>('/api/inventory/stock?skus=' + skus.join(','));
	return data.items ?? [];
}
