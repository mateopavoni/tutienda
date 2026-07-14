import { error } from '@sveltejs/kit';
import { enabledPages, pageSlug } from '$lib/pages';
import type { PageLoad } from './$types';

// A standalone storefront page (just Terms — about/faq/contact are the platform's own marketing pages
// now, not per-store ones). The store (with its settings) is already loaded
// by the [slug] layout; we just pick the matching page here, by its route slug (auto-derived from the
// title server-side, but editable — see $lib/pages). An unknown segment or a page with no content 404s —
// this dynamic route sits next to the static `product/` route, which keeps taking precedence.
export const load: PageLoad = async ({ parent, params }) => {
	const { store } = await parent();
	const page = enabledPages(store?.settings?.pages).find((p) => pageSlug(p) === params.page);
	if (!page) throw error(404, `Page "${params.page}" not found`);
	return { page };
};
