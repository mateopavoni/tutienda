// Contrast guards for the per-store custom colors (accent, title). Both take a merchant-chosen #rrggbb
// hex and the resolved light/dark mode, and hand back a "r g b" triplet for a CSS var (or undefined to
// fall back to the design system's own default) — never an unreadable custom color.

function parseHex(hex: string | undefined): { r: number; g: number; b: number; lum: number } | null {
	if (!hex) return null;
	const m = /^#?([0-9a-fA-F]{6})$/.exec(hex.trim());
	if (!m) return null;
	const n = parseInt(m[1], 16);
	const r = (n >> 16) & 255;
	const g = (n >> 8) & 255;
	const b = n & 255;
	const lum = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
	return { r, g, b, lum };
}

// safeAccent guards a color used as a button fill (white text on top): too light and the text vanishes.
// In dark mode the same accent also renders as plain foreground text over the app's dark surfaces, so a
// very dark accent (chocolate brown, forest green) is lightened towards white until it's legible there.
export function safeAccent(hex: string | undefined, dark: boolean): string | undefined {
	const c = parseHex(hex);
	if (!c) return undefined;
	let { r, g, b } = c;
	if (c.lum > 0.7) return undefined;
	if (dark && c.lum < 0.4) {
		const mix = (v: number) => Math.round(v + (255 - v) * 0.5);
		r = mix(r);
		g = mix(g);
		b = mix(b);
	}
	return `${r} ${g} ${b}`;
}

// safeTitleColor guards a color used as plain heading text directly on the page background (bg-bg, which
// is near-white in light mode and near-black in dark mode) — the opposite failure mode from an accent
// button fill. A color that would nearly disappear against that background falls back to undefined (the
// design system's own text-text) instead of a half-guessed correction.
export function safeTitleColor(hex: string | undefined, dark: boolean): string | undefined {
	const c = parseHex(hex);
	if (!c) return undefined;
	if (!dark && c.lum > 0.85) return undefined; // near-white text would vanish on the light page background
	if (dark && c.lum < 0.15) return undefined; // near-black text would vanish on the dark page background
	return `${c.r} ${c.g} ${c.b}`;
}
