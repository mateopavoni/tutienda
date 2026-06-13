// Storefront layout model. A store is composed of ordered sections; some are fixed across every store
// (the nav/footer chrome and the design tokens never change — that's the "Monolith Editorial" system),
// while these sections are the part a merchant arranges: which appear, in what order, and their copy.
// The product grid ("catalog") is mandatory and always rendered. Mirrors accounts.Settings.Layout in Go.

export type SectionType =
	| 'announcement'
	| 'hero'
	| 'feature'
	| 'catalog'
	| 'about'
	| 'contact'
	| 'gallery'
	| 'faq';

// ContentField is the set of editable text knobs on a section (everything except its type/enabled flag).
// The editor form iterates these, so typing them precisely keeps section[field] a string, not a boolean.
export type ContentField = 'heading' | 'body' | 'imageUrl' | 'ctaLabel';

// SectionItem is one repeated entry inside an item-based section (gallery image or FAQ Q&A). Mirrors
// accounts.SectionItem in Go.
export interface SectionItem {
	heading?: string;
	body?: string;
	imageUrl?: string;
}

export interface Section {
	type: SectionType;
	enabled: boolean;
	heading?: string;
	body?: string;
	imageUrl?: string;
	ctaLabel?: string;
	items?: SectionItem[];
}

// ItemKind tells the editor/storefront how to treat a section's repeatable items: 'image' = a gallery
// image (imageUrl + optional caption), 'qa' = an FAQ entry (heading question + body answer).
export type ItemKind = 'image' | 'qa';

// Which content fields each section type uses (drives the editor form) and whether it's fixed (the grid
// can be reordered but never disabled/removed). The canonical order is also the order missing sections
// are appended in.
export const SECTION_META: Record<
	SectionType,
	{ label: string; fields: ContentField[]; fixed?: boolean; hint: string; item?: { kind: ItemKind; label: string } }
> = {
	announcement: { label: 'Announcement bar', fields: ['body'], hint: 'A thin banner across the very top.' },
	hero: {
		label: 'Hero',
		fields: ['heading', 'body', 'imageUrl', 'ctaLabel'],
		hint: 'A large banner with a headline, image and a button to your catalog.'
	},
	feature: {
		label: 'Feature band',
		fields: ['heading', 'body', 'imageUrl', 'ctaLabel'],
		hint: 'A side-by-side image and text block — great for a promo or a story.'
	},
	catalog: { label: 'Product grid', fields: [], fixed: true, hint: 'Your products. Always shown.' },
	about: { label: 'About', fields: ['heading', 'body'], hint: 'An editorial text block below the grid.' },
	contact: {
		label: 'Contact',
		fields: ['heading', 'body'],
		hint: 'Address, hours or how to reach you, below the grid.'
	},
	gallery: {
		label: 'Gallery',
		fields: ['heading'],
		hint: 'A grid of images — lookbook, space, or work.',
		item: { kind: 'image', label: 'Image' }
	},
	faq: {
		label: 'FAQ',
		fields: ['heading'],
		hint: 'A list of questions and answers.',
		item: { kind: 'qa', label: 'Question' }
	}
};

const SECTION_TYPES = Object.keys(SECTION_META) as SectionType[];

function isSectionType(t: unknown): t is SectionType {
	return typeof t === 'string' && (SECTION_TYPES as string[]).includes(t);
}

/** defaultSection returns a section of a type with sensible starter copy (off by default, except the grid). */
export function defaultSection(type: SectionType): Section {
	switch (type) {
		case 'announcement':
			return { type, enabled: false, body: 'Free shipping on orders over $100' };
		case 'hero':
			return {
				type,
				enabled: false,
				heading: 'New season',
				body: 'The latest drop, now live.',
				imageUrl: '',
				ctaLabel: 'Shop now'
			};
		case 'feature':
			return {
				type,
				enabled: false,
				heading: 'Crafted with care',
				body: 'Tell the story behind your products.',
				imageUrl: '',
				ctaLabel: ''
			};
		case 'about':
			return { type, enabled: false, heading: 'About', body: '' };
		case 'contact':
			return { type, enabled: false, heading: 'Visit us', body: '' };
		case 'gallery':
			return { type, enabled: false, heading: 'Gallery', items: [] };
		case 'faq':
			return { type, enabled: false, heading: 'FAQ', items: [{ heading: '', body: '' }] };
		case 'catalog':
			return { type, enabled: true };
	}
}

/** DEFAULT_LAYOUT is the full, ordered editor model for a store that hasn't customized anything. */
export const DEFAULT_LAYOUT: Section[] = SECTION_TYPES.map(defaultSection);

/**
 * normalizeLayout returns every known section exactly once, preserving the saved order for sections that
 * were stored and appending any missing ones in canonical order. The grid is forced enabled (mandatory),
 * and unknown/legacy types are dropped. Both the storefront and the editor work off this shape.
 */
export function normalizeLayout(raw?: Section[] | null): Section[] {
	const saved = new Map<SectionType, Section>();
	const order: SectionType[] = [];
	for (const s of raw ?? []) {
		if (!isSectionType(s?.type) || saved.has(s.type)) continue;
		saved.set(s.type, { ...defaultSection(s.type), ...s, type: s.type });
		order.push(s.type);
	}
	for (const t of SECTION_TYPES) if (!order.includes(t)) order.push(t);
	return order.map((t) => {
		const s = saved.get(t) ?? defaultSection(t);
		return t === 'catalog' ? { ...s, enabled: true } : s;
	});
}

/** enabledSections is what the storefront actually renders, in order. */
export function enabledSections(raw?: Section[] | null): Section[] {
	return normalizeLayout(raw).filter((s) => s.enabled);
}
