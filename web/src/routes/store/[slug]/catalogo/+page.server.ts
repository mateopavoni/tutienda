import type { PageServerLoad } from './$types';
import { listProducts } from '$lib/server/api';

// Dedicated catalog page: fetch the store's full catalog so it can build its category filter from the
// store's *own* products (every store sells something different). The grid then filters in-page.
export const load: PageServerLoad = async ({ fetch, url, parent }) => {
	const { storeSlug } = await parent();
	const category = url.searchParams.get('category') ?? '';
	try {
		const products = await listProducts(fetch, storeSlug);
		return { products, category, failed: false };
	} catch {
		return { products: [], category, failed: true };
	}
};
