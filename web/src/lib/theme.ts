// Storefront visual themes. The truth (the closed set + default) lives in Go (accounts.knownThemes /
// normalizeTheme); this mirror drives the /app picker and lets the storefront resolve a safe id. The
// actual tokens (fonts + palette) live in app.css under [data-store-theme]; here we only carry id +
// labels for the UI. Keep THEME_IDS in sync with knownThemes in services/internal/accounts/service.go.

export const DEFAULT_THEME = 'monolith';

export type ThemeId = 'monolith' | 'boutique' | 'pop' | 'terminal';

export interface ThemeMeta {
	id: ThemeId;
	label: string;
	description: string;
}

export const THEMES: ThemeMeta[] = [
	{ id: 'monolith', label: 'Monolith', description: 'Brutalist editorial — Inter, hard edges, high contrast.' },
	{ id: 'boutique', label: 'Boutique', description: 'Warm serif (Fraunces), cream & ink, terracotta accent.' },
	{ id: 'pop', label: 'Pop', description: 'Bright rounded sans (Poppins), bold colour.' },
	{ id: 'terminal', label: 'Terminal', description: 'Full monospace, green-on-paper, technical.' }
];

const KNOWN = new Set<string>(THEMES.map((t) => t.id));

// normalizeTheme keeps a known theme, else returns the default — mirrors Go's normalizeTheme so a tampered
// or legacy value never reaches a [data-store-theme] block that doesn't exist.
export function normalizeTheme(theme: string | undefined | null): ThemeId {
	const t = (theme ?? '').trim();
	return (KNOWN.has(t) ? t : DEFAULT_THEME) as ThemeId;
}
