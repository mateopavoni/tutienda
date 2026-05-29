import { describe, it, expect } from 'vitest';
import { get } from 'svelte/store';
import { t, locale } from './index';

describe('i18n', () => {
	it('translates a key in the active locale', () => {
		locale.set('en');
		expect(get(t)('cart.title')).toBe('Cart');
		locale.set('es');
		expect(get(t)('cart.title')).toBe('Carrito');
		locale.set('pt-BR');
		expect(get(t)('cart.title')).toBe('Carrinho');
	});

	it('interpolates parameters', () => {
		locale.set('en');
		expect(get(t)('common.lowStock', { n: 3 })).toBe('Only 3 left');
	});

	it('falls back to the key when missing', () => {
		locale.set('en');
		expect(get(t)('does.not.exist')).toBe('does.not.exist');
	});
});
