// Browser client for storefront customer auth + order history. Talks to the public gateway, which
// resolves the store from X-Tenant-Slug (signup/login) or the buyer token's own tenant claim (orders).
import { env } from '$env/dynamic/public';
import type { ApiError, Order } from '$lib/types';
import { getBrowserSlug } from '$lib/tenant';
import type { CustomerProfile } from '$lib/stores/customer';

function base(): string {
	return env.PUBLIC_API_BASE || 'http://localhost:8080';
}

export interface CustomerSession {
	token: string;
	customer: CustomerProfile;
}

async function call<T>(path: string, init: RequestInit, token?: string): Promise<T> {
	const headers = new Headers(init.headers);
	headers.set('Content-Type', 'application/json');
	headers.set('X-Tenant-Slug', getBrowserSlug());
	if (token) headers.set('Authorization', 'Bearer ' + token);
	const res = await fetch(base() + path, { ...init, headers });
	if (res.status === 204) return undefined as T;
	const data = (await res.json().catch(() => ({}))) as unknown;
	if (!res.ok) throw { ...(data as ApiError), status: res.status };
	return data as T;
}

export async function customerSignup(email: string, password: string, name: string): Promise<CustomerSession> {
	return call<CustomerSession>('/api/customers/signup', {
		method: 'POST',
		body: JSON.stringify({ email, password, name })
	});
}

export async function customerLogin(email: string, password: string): Promise<CustomerSession> {
	return call<CustomerSession>('/api/customers/login', {
		method: 'POST',
		body: JSON.stringify({ email, password })
	});
}

// myOrders returns the signed-in customer's order history (newest first).
export async function myOrders(token: string): Promise<Order[]> {
	const data = await call<{ items: Order[] }>('/api/orders/mine', { method: 'GET' }, token);
	return data.items ?? [];
}
