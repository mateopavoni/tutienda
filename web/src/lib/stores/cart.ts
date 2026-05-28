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

const STORAGE_KEY = 'cart';

function initial(): CartLine[] {
	if (!browser) return [];
	try {
		return JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]') as CartLine[];
	} catch {
		return [];
	}
}

function createCart() {
	const { subscribe, update, set } = writable<CartLine[]>(initial());
	if (browser) {
		subscribe((lines) => localStorage.setItem(STORAGE_KEY, JSON.stringify(lines)));
	}

	return {
		subscribe,
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
