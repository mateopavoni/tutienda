import type { PageServerLoad } from './$types';
import { listProducts } from '$lib/server/api';

export const load: PageServerLoad = async ({ fetch, url, parent }) => {
	const { storeSlug } = await parent();
	const category = url.searchParams.get('category') ?? '';
	try {
		// Fetch the full catalog (not just the selected category) so the storefront can build its category
		// filter from the store's *own* products — every store sells something different, so a hardcoded
		// category list would be wrong for any store but the original demo. The grid then filters in-page.
		const products = await listProducts(fetch, storeSlug);
		return { products, category, failed: false };
	} catch {
		return { products: [], category, failed: true };
	}
};
