// Tenant (store) resolution for the storefront. The store is scoped by the /store/[slug] route param
// (see +layout.server.ts); the resolved slug is sent to the gateway as X-Tenant-Slug, which maps it to
// the tenant id every backend service is scoped to.

import { env } from '$env/dynamic/public';

export const DEFAULT_SLUG = 'system-archive';

// Browser-side holder: the layout load resolves the slug server-side and seeds it here so the
// browser API client (cart/checkout/live stock) can send the same X-Tenant-Slug.
let browserSlug = '';
export function setBrowserSlug(slug: string): void {
	browserSlug = slug;
}
export function getBrowserSlug(): string {
	return browserSlug || env.PUBLIC_STORE_SLUG || DEFAULT_SLUG;
}

/** Store branding the storefront applies on top of the fixed design system. */
export interface StoreInfo {
	id: string;
	slug: string;
	displayName: string;
	settings: {
		logoUrl?: string;
		accentColor?: string;
		theme?: string;
		currency?: string;
		layout?: import('$lib/layout').Section[];
		pages?: import('$lib/pages').Page[];
	};
}
