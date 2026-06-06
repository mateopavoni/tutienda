import type { LayoutServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { getStore } from '$lib/server/api';

// The storefront is scoped by the [slug] route param (in production a subdomain rewrites to this path).
// We resolve the store once per navigation and hand it to every child page via parent(); the slug also
// seeds the browser API client so client-side calls (cart/checkout/live stock) stay scoped to this store.
export const load: LayoutServerLoad = async ({ fetch, params }) => {
	const storeSlug = params.slug;
	try {
		const store = await getStore(fetch, storeSlug);
		// getStore returns null only on a 404: the slug maps to no store → a real not-found.
		if (store === null) throw error(404, `Store "${storeSlug}" not found`);
		return { storeSlug, store };
	} catch (e) {
		// Re-throw the deliberate 404; swallow transient backend failures so the page renders degraded
		// (with defaults) instead of 500-ing on a flaky gateway/accounts.
		if (e && typeof e === 'object' && 'status' in e) throw e;
		return { storeSlug, store: null };
	}
};
