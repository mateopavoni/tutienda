// Tenant (store) resolution for the storefront. A deployment can serve many stores: in production each
// store lives on its own subdomain (acme.example.com → "acme"); in local dev there are no subdomains,
// so we fall back to PUBLIC_STORE_SLUG (defaulting to the seeded demo store). The resolved slug is sent
// to the gateway as X-Tenant-Slug, which maps it to the tenant id every backend service is scoped to.

import { env } from '$env/dynamic/public';

export const DEFAULT_SLUG = 'system-archive';

/** slugFromHost returns the leftmost label of a host when it looks like slug.domain.tld, else ''. */
export function slugFromHost(host: string | null | undefined): string {
	if (!host) return '';
	const h = host.split(':')[0];
	if (h === '' || h === 'localhost') return '';
	const labels = h.split('.');
	if (labels.length < 3) return '';
	const first = labels[0];
	if (first === 'www' || first === 'api' || first === 'admin') return '';
	return first;
}

/** resolveSlug picks the store slug for a request: subdomain first, then the configured default. */
export function resolveSlug(host: string | null | undefined): string {
	return slugFromHost(host) || env.PUBLIC_STORE_SLUG || DEFAULT_SLUG;
}

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
		currency?: string;
		layout?: import('$lib/layout').Section[];
		pages?: import('$lib/pages').Page[];
	};
}
