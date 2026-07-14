import { describe, it, expect } from 'vitest';
import { enabledPages, pageTitle, pageSlug, type Page } from './pages';

describe('pages mirror (matches Go sanitizePages)', () => {
	it('keeps only content pages, dropping a blank one', () => {
		const pages: Page[] = [{ type: 'terms', title: 'Terms', body: 'no returns' }];
		const out = enabledPages(pages);
		expect(out.map((p) => p.type)).toEqual(['terms']);
		expect(out[0].title).toBe('Terms');
	});

	it('returns [] for undefined / all-empty', () => {
		expect(enabledPages(undefined)).toEqual([]);
		expect(enabledPages([{ type: 'terms', body: '' }])).toEqual([]);
	});

	it('pageTitle falls back to the type default when title is blank', () => {
		expect(pageTitle({ type: 'terms', title: '  ', body: 'x' })).toBe('Terms');
		expect(pageTitle({ type: 'terms', title: 'Terms & Conditions', body: 'x' })).toBe('Terms & Conditions');
	});

	it('pageSlug falls back to the type when no slug was set (e.g. an unsaved row)', () => {
		expect(pageSlug({ type: 'terms', slug: 'condiciones', body: 'x' })).toBe('condiciones');
		expect(pageSlug({ type: 'terms', slug: '  ', body: 'x' })).toBe('terms');
	});
});
