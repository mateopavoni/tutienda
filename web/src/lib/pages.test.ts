import { describe, it, expect } from 'vitest';
import { enabledPages, pageTitle, type Page } from './pages';

describe('pages mirror (matches Go sanitizePages)', () => {
	it('keeps only content pages, in canonical order, deduped', () => {
		const pages: Page[] = [
			{ type: 'terms', title: 'Terms', body: 'no returns' },
			{ type: 'about', title: 'About us', body: 'hi' },
			{ type: 'faq', title: 'FAQ', body: '   ' }, // blank body → absent
			{ type: 'about', title: 'dup', body: 'dup' } // duplicate → ignored
		];
		const out = enabledPages(pages);
		expect(out.map((p) => p.type)).toEqual(['about', 'terms']); // canonical order, faq dropped
		expect(out[0].title).toBe('About us'); // first about wins
	});

	it('returns [] for undefined / all-empty', () => {
		expect(enabledPages(undefined)).toEqual([]);
		expect(enabledPages([{ type: 'about', body: '' }])).toEqual([]);
	});

	it('pageTitle falls back to the type default when title is blank', () => {
		expect(pageTitle({ type: 'faq', title: '  ', body: 'x' })).toBe('FAQ');
		expect(pageTitle({ type: 'about', title: 'Our story', body: 'x' })).toBe('Our story');
	});
});
