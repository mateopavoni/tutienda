import { error } from '@sveltejs/kit';
import { enabledPages, type PageType } from '$lib/pages';
import type { PageLoad } from './$types';

// A standalone storefront page (about/faq/contact/terms). The store (with its settings) is already loaded
// by the [slug] layout; we just pick the matching page here. An unknown segment or a page with no content
// 404s — this dynamic route sits next to the static `product/` route, which keeps taking precedence.
export const load: PageLoad = async ({ parent, params }) => {
	const { store } = await parent();
	const page = enabledPages(store?.settings?.pages).find((p) => p.type === (params.page as PageType));
	if (!page) throw error(404, `Page "${params.page}" not found`);
	return { page };
};
