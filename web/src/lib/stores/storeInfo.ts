import { writable } from 'svelte/store';
import { env } from '$env/dynamic/public';
import type { StoreInfo } from '$lib/tenant';

// Live mirror of the SSR-resolved store: seeded from the layout load, then refreshed by a background
// poll (see +layout.svelte) so a merchant's saved branding/layout changes reach an already-open
// storefront tab without the visitor reloading. null until the layout seeds it.
export const storeInfo = writable<StoreInfo | null>(null);

function base(): string {
	return env.PUBLIC_API_BASE || 'http://localhost:8080';
}

// fetchStoreInfo hits the public by-slug lookup directly (no tenant header needed: the slug is in the
// path), the same endpoint the SSR load uses.
export async function fetchStoreInfo(slug: string): Promise<StoreInfo | null> {
	const res = await fetch(base() + '/api/accounts/stores/by-slug/' + encodeURIComponent(slug));
	if (!res.ok) return null;
	return (await res.json()) as StoreInfo;
}
