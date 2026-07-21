import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { getProduct, getStock } from '$lib/server/api';

export const load: PageServerLoad = async ({ params, fetch }) => {
	const storeSlug = params.slug;
	const product = await getProduct(fetch, storeSlug, params.id);
	if (!product) {
		throw error(404, 'Product not found');
	}

	const stock = await getStock(
		fetch,
		storeSlug,
		product.variants.map((v) => v.sku)
	).catch(() => []);
	const stockMap: Record<string, number> = {};
	for (const s of stock) {
		stockMap[s.sku] = s.available;
	}

	return { product, stockMap };
};
