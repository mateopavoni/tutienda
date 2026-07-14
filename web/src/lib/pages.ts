// Standalone storefront pages, each at its own route (/store/{slug}/{slug-or-type}). Unlike layout
// sections (homepage blocks in $lib/layout), these are full pages the merchant fills in. Mirrors
// accounts.Page / sanitizePages in Go: the closed set of types and the "empty body = absent" rule must
// agree both sides.

export type PageType = 'terms';

export interface Page {
	type: PageType;
	/** Route segment, auto-derived server-side from title (falls back to type) but editable. */
	slug?: string;
	title?: string;
	body?: string;
}

// pageSlug is the actual route segment for a page — its own slug, or the type if none was set yet (e.g.
// a brand-new row in the /app editor that hasn't been saved). Mirrors accounts.pageSlug in Go.
export function pageSlug(p: Page): string {
	return (p.slug ?? '').trim() || p.type;
}

// Canonical order + per-type defaults. The order drives the editor and the footer links; defaultTitle is
// used when the merchant left the title blank.
export const PAGE_META: Record<PageType, { defaultTitle: string; hint: string }> = {
	terms: { defaultTitle: 'Terms', hint: 'Terms, returns & policies.' }
};

export const PAGE_TYPES: PageType[] = ['terms'];

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
