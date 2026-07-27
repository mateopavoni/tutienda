// Browser API client for the merchant admin panel. Talks to the public gateway: accounts endpoints use
// the merchant token, admin catalog/inventory endpoints use the store-scoped token.
import { env } from '$env/dynamic/public';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import type { ApiError, Message, Product, Stock } from '$lib/types';
import type { Section } from '$lib/layout';
import type { Page } from '$lib/pages';
import { session, type ActiveStore } from './session';

function base(): string {
	return env.PUBLIC_API_BASE || 'http://localhost:8080';
}

export interface Merchant {
	id: string;
	email: string;
	/** Platform role: 'merchant' (default) or 'admin' (SaaS super-admin). */
	role: string;
	createdAt?: string;
}

export interface PlatformStats {
	merchants: number;
	stores: number;
	storesByPlan: Record<string, number>;
}
export interface Store {
	id: string;
	slug: string;
	ownerId: string;
	displayName: string;
	settings: {
		logoUrl?: string;
		accentColor?: string;
		titleColor?: string;
		theme?: string;
		currency?: string;
		layout?: Section[];
		pages?: Page[];
	};
	/** Membership tier (see $lib/plan). Defaults to 'free' for stores created before plans existed. */
	plan: string;
	/** When true, the storefront is switched off (404s at the gateway) without deleting the store. */
	disabled: boolean;
}
interface Session {
	token: string;
	merchant?: Merchant;
	store?: Store;
}

async function call<T>(path: string, init: RequestInit, token?: string, retried = false): Promise<T> {
	const headers = new Headers(init.headers);
	headers.set('Content-Type', 'application/json');
	if (token) headers.set('Authorization', 'Bearer ' + token);
	const res = await fetch(base() + path, { ...init, headers });
	// A 401 on a call that carried a token means the session died server-side (expiry, not a bad
	// login attempt — those never pass a token). The store token is the one that expires mid-session
	// most often (every /app write uses it); if the merchant token is still alive, re-mint the store
	// token once (same endpoint as selectStore) instead of nuking the whole session over a token the
	// user didn't even know had a separate lifetime. Only bounce to /login when that recovery isn't
	// possible (merchant token itself expired, or the retry still 401s).
	if (res.status === 401 && token) {
		const merchantToken = session.merchantToken();
		const active = session.activeStore();
		if (!retried && token === session.storeToken() && merchantToken && active) {
			try {
				const fresh = await call<{ token: string }>(
					'/api/accounts/stores/' + encodeURIComponent(active.id) + '/token',
					{ method: 'POST' },
					merchantToken
				);
				session.setStore(fresh.token, active);
				return call<T>(path, init, fresh.token, true);
			} catch {
				// merchant token is also dead — fall through to the hard logout below
			}
		}
		session.clearAll();
		if (browser) goto('/login?expired=1');
		throw { error: 'session expired' } as ApiError;
	}
	if (res.status === 204) return undefined as T;
	const data = (await res.json().catch(() => ({}))) as unknown;
	if (!res.ok) throw data as ApiError;
	return data as T;
}

// --- Auth (merchant token) ---

export async function signup(email: string, password: string): Promise<Session> {
	const s = await call<Session>('/api/accounts/signup', {
		method: 'POST',
		body: JSON.stringify({ email, password })
	});
	session.setMerchantToken(s.token);
	session.setRole(s.merchant?.role ?? 'merchant');
	return s;
}

export async function login(email: string, password: string): Promise<Session> {
	const s = await call<Session>('/api/accounts/login', {
		method: 'POST',
		body: JSON.stringify({ email, password })
	});
	session.setMerchantToken(s.token);
	session.setRole(s.merchant?.role ?? 'merchant');
	return s;
}

// --- Stores (merchant token) ---

export async function listStores(): Promise<Store[]> {
	const data = await call<{ items: Store[] }>('/api/accounts/stores', { method: 'GET' }, session.merchantToken());
	return data.items ?? [];
}

export async function createStore(
	slug: string,
	displayName: string,
	theme?: string,
	accentColor?: string
): Promise<Store> {
	return call<Store>(
		'/api/accounts/stores',
		{ method: 'POST', body: JSON.stringify({ slug, displayName, theme, accentColor }) },
		session.merchantToken()
	);
}

export async function updateStore(id: string, displayName: string, settings: Store['settings']): Promise<Store> {
	return call<Store>(
		'/api/accounts/stores/' + encodeURIComponent(id),
		{ method: 'PATCH', body: JSON.stringify({ displayName, settings }) },
		session.merchantToken()
	);
}

// changePlan switches a store's membership tier. The new tier is signed into the *store* token, so the
// caller must re-mint it (selectStore) afterwards for the gateway to honor the upgrade.
export async function changePlan(id: string, plan: string): Promise<Store> {
	return call<Store>(
		'/api/accounts/stores/' + encodeURIComponent(id) + '/plan',
		{ method: 'PATCH', body: JSON.stringify({ plan }) },
		session.merchantToken()
	);
}

// deleteStore permanently removes a store and cascades to its products/stock. Irreversible.
export async function deleteStore(id: string): Promise<void> {
	await call<void>('/api/accounts/stores/' + encodeURIComponent(id), { method: 'DELETE' }, session.merchantToken());
}

// setStoreDisabled toggles a store's storefront on/off.
export async function setStoreDisabled(id: string, disabled: boolean): Promise<Store> {
	return call<Store>(
		'/api/accounts/stores/' + encodeURIComponent(id) + '/status',
		{ method: 'PATCH', body: JSON.stringify({ disabled }) },
		session.merchantToken()
	);
}

// selectStore mints a store-scoped token and remembers the active store.
export async function selectStore(store: Store): Promise<void> {
	const s = await call<Session>(
		'/api/accounts/stores/' + encodeURIComponent(store.id) + '/token',
		{ method: 'POST' },
		session.merchantToken()
	);
	const active: ActiveStore = { id: store.id, slug: store.slug, displayName: store.displayName };
	session.setStore(s.token, active);
}

export interface ImportRowError {
	row: number;
	message: string;
}
export interface ImportResult {
	created: number;
	errors: ImportRowError[];
}

// importProducts bulk-creates products from a CSV file. Same non-`call()` multipart approach as
// uploadImage (the browser must set its own Content-Type boundary).
export async function importProducts(file: File): Promise<ImportResult> {
	const form = new FormData();
	form.append('file', file);
	const headers = new Headers();
	const token = session.storeToken();
	if (token) headers.set('Authorization', 'Bearer ' + token);
	const res = await fetch(base() + '/api/admin/catalog/products/import', { method: 'POST', body: form, headers });
	const data = (await res.json().catch(() => ({}))) as Partial<ImportResult> & ApiError;
	if (!res.ok) throw data as ApiError;
	return { created: data.created ?? 0, errors: data.errors ?? [] };
}

// listMessages returns a store's contact-form submissions, newest first (merchant token, ownership
// checked server-side — no store token needed since this doesn't touch catalog/inventory).
export async function listMessages(storeId: string): Promise<Message[]> {
	const data = await call<{ items: Message[] }>(
		'/api/accounts/stores/' + encodeURIComponent(storeId) + '/messages',
		{ method: 'GET' },
		session.merchantToken()
	);
	return data.items ?? [];
}

// --- Catalog + inventory (store token) ---

export async function listProducts(): Promise<Product[]> {
	const data = await call<{ items: Product[] }>('/api/admin/catalog/products', { method: 'GET' }, session.storeToken());
	return data.items ?? [];
}

export async function createProduct(p: Partial<Product>): Promise<Product> {
	return call<Product>('/api/admin/catalog/products', { method: 'POST', body: JSON.stringify(p) }, session.storeToken());
}

export async function updateProduct(id: string, p: Partial<Product>): Promise<Product> {
	return call<Product>(
		'/api/admin/catalog/products/' + encodeURIComponent(id),
		{ method: 'PUT', body: JSON.stringify(p) },
		session.storeToken()
	);
}

export async function deleteProduct(id: string): Promise<void> {
	await call<void>('/api/admin/catalog/products/' + encodeURIComponent(id), { method: 'DELETE' }, session.storeToken());
}

export async function setStock(sku: string, available: number): Promise<void> {
	await call<void>(
		'/api/admin/inventory/stock',
		{ method: 'POST', body: JSON.stringify({ sku, available }) },
		session.storeToken()
	);
}

// listStock reads current on-hand for a set of SKUs (store token → tenant-scoped), so the dashboard can
// show stock per variant inline. Goes through the admin inventory proxy, same as setStock.
export async function listStock(skus: string[]): Promise<Stock[]> {
	if (skus.length === 0) return [];
	const data = await call<{ items: Stock[] }>(
		'/api/admin/inventory/stock?skus=' + encodeURIComponent(skus.join(',')),
		{ method: 'GET' },
		session.storeToken()
	);
	return data.items ?? [];
}

// uploadImage sends a multipart file to the object store via catalog and returns its public URL. It does
// not use `call`: multipart must NOT set Content-Type by hand (the browser adds the boundary), so this
// builds the request directly. Used for product photos and section images.
export async function uploadImage(file: File): Promise<string> {
	const form = new FormData();
	form.append('file', file);
	const headers = new Headers();
	const token = session.storeToken();
	if (token) headers.set('Authorization', 'Bearer ' + token);
	const res = await fetch(base() + '/api/admin/catalog/uploads', { method: 'POST', body: form, headers });
	const data = (await res.json().catch(() => ({}))) as { url?: string } & ApiError;
	if (!res.ok) throw data as ApiError;
	return data.url ?? '';
}

// --- Platform super-admin (merchant token with the admin role; gateway gates /api/platform) ---

export async function platformStats(): Promise<PlatformStats> {
	return call<PlatformStats>('/api/platform/stats', { method: 'GET' }, session.merchantToken());
}

export async function platformStores(): Promise<Store[]> {
	const data = await call<{ items: Store[] }>('/api/platform/stores', { method: 'GET' }, session.merchantToken());
	return data.items ?? [];
}

export async function platformMerchants(): Promise<Merchant[]> {
	const data = await call<{ items: Merchant[] }>('/api/platform/merchants', { method: 'GET' }, session.merchantToken());
	return data.items ?? [];
}

export async function platformSetPlan(storeId: string, plan: string): Promise<Store> {
	return call<Store>(
		'/api/platform/stores/' + encodeURIComponent(storeId) + '/plan',
		{ method: 'PATCH', body: JSON.stringify({ plan }) },
		session.merchantToken()
	);
}
