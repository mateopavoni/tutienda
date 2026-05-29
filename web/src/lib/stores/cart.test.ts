import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { cart, cartCount, cartTotal } from './cart';

const sample = {
	sku: 'AX1-85',
	productId: 'aero-strata-x1',
	name: 'Aero Strata_X1',
	label: 'US 8.5',
	priceCents: 28500,
	currency: 'USD',
	imageUrl: 'x'
};

describe('cart store', () => {
	beforeEach(() => cart.clear());

	it('adds a line and counts quantity', () => {
		cart.add(sample);
		expect(get(cartCount)).toBe(1);
		expect(get(cartTotal)).toBe(28500);
	});

	it('merges quantity for the same sku instead of duplicating', () => {
		cart.add(sample);
		cart.add(sample, 2);
		expect(get(cart).length).toBe(1);
		expect(get(cartCount)).toBe(3);
		expect(get(cartTotal)).toBe(85500);
	});

	it('removes a line when quantity drops to zero', () => {
		cart.add(sample);
		cart.setQty('AX1-85', 0);
		expect(get(cart).length).toBe(0);
	});

	it('clears the cart', () => {
		cart.add(sample);
		cart.clear();
		expect(get(cartCount)).toBe(0);
	});
});
