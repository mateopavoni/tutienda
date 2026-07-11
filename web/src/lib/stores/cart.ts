import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

export interface CartLine {
	sku: string;
	productId: string;
	name: string;
	label: string;
	priceCents: number;
	currency: string;
	imageUrl: string;
	qty: number;
}

// Each store gets its own cart: localStorage is per-origin, not per-path, so without a per-slug key
// every /store/[slug] would share one cart (a real cross-tenant bug — a buyer would see another
// merchant's demo items in their bag).
function storageKey(slug: string): string {
	return `cart:${slug}`;
}

function readLines(slug: string): CartLine[] {
	if (!browser || !slug) return [];
	try {
		return JSON.parse(localStorage.getItem(storageKey(slug)) ?? '[]') as CartLine[];
	} catch {
		return [];
	}
}

function createCart() {
	let currentSlug = '';
	const { subscribe, update, set } = writable<CartLine[]>([]);
	if (browser) {
		subscribe((lines) => {
			if (currentSlug) localStorage.setItem(storageKey(currentSlug), JSON.stringify(lines));
		});
	}

	return {
		subscribe,
		// Called once the storefront layout resolves the slug (and again if it changes, e.g. navigating
		// between two stores in the same tab): swaps in that store's own cart from localStorage.
		hydrate(slug: string) {
			if (!slug || slug === currentSlug) return;
			currentSlug = slug;
			set(readLines(slug));
		},
		add(line: Omit<CartLine, 'qty'>, qty = 1) {
			update((lines) => {
				const existing = lines.find((l) => l.sku === line.sku);
				if (existing) {
					return lines.map((l) => (l.sku === line.sku ? { ...l, qty: l.qty + qty } : l));
				}
				return [...lines, { ...line, qty }];
			});
		},
		setQty(sku: string, qty: number) {
			update((lines) =>
				qty <= 0 ? lines.filter((l) => l.sku !== sku) : lines.map((l) => (l.sku === sku ? { ...l, qty } : l))
			);
		},
		remove(sku: string) {
			update((lines) => lines.filter((l) => l.sku !== sku));
		},
		clear() {
			set([]);
		}
	};
}

export const cart = createCart();
export const cartCount = derived(cart, ($cart) => $cart.reduce((sum, l) => sum + l.qty, 0));
export const cartTotal = derived(cart, ($cart) => $cart.reduce((sum, l) => sum + l.priceCents * l.qty, 0));
export const cartCurrency = derived(cart, ($cart) => $cart[0]?.currency ?? 'USD');
