// Standalone storefront pages, each at its own route (/store/{slug}/{type}). Unlike layout sections
// (homepage blocks in $lib/layout), these are full pages the merchant fills in. Mirrors accounts.Page /
// sanitizePages in Go: the closed set of types and the "empty body = absent" rule must agree both sides.

export type PageType = 'about' | 'faq' | 'contact' | 'terms';

export interface Page {
	type: PageType;
	title?: string;
	body?: string;
}

// Canonical order + per-type defaults. The order drives the editor and the footer links; defaultTitle is
// used when the merchant left the title blank.
export const PAGE_META: Record<PageType, { defaultTitle: string; hint: string }> = {
	about: { defaultTitle: 'About', hint: 'Your story / brand.' },
	faq: { defaultTitle: 'FAQ', hint: 'Frequently asked questions.' },
	contact: { defaultTitle: 'Contact', hint: 'How customers reach you.' },
	terms: { defaultTitle: 'Terms', hint: 'Terms, returns & policies.' }
};

export const PAGE_TYPES: PageType[] = ['about', 'faq', 'contact', 'terms'];

// pageTitle is the displayed heading: the merchant's title, or the type default when blank.
export function pageTitle(p: Page): string {
	return (p.title ?? '').trim() || PAGE_META[p.type].defaultTitle;
}

// enabledPages returns the pages that have content, in canonical order — what the storefront routes and
// the footer links to. Mirrors the server rule that an empty-body page is absent.
export function enabledPages(pages: Page[] | undefined): Page[] {
	const byType = new Map<PageType, Page>();
	for (const p of pages ?? []) {
		if (PAGE_TYPES.includes(p.type) && (p.body ?? '').trim() && !byType.has(p.type)) byType.set(p.type, p);
	}
	return PAGE_TYPES.filter((t) => byType.has(t)).map((t) => byType.get(t)!);
}
