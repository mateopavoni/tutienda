import { describe, it, expect } from 'vitest';
import { normalizeTheme, DEFAULT_THEME, THEMES } from './theme';

// Mirrors normalizeTheme in services/internal/accounts/service.go. Unknown/tampered ⇒ default theme,
// same defensive stance as an unknown plan tier falling to free.
describe('theme mirror (matches Go normalizeTheme)', () => {
	it('falls back to the default for unknown/empty/null', () => {
		expect(normalizeTheme(undefined)).toBe(DEFAULT_THEME);
		expect(normalizeTheme(null)).toBe(DEFAULT_THEME);
		expect(normalizeTheme('')).toBe(DEFAULT_THEME);
		expect(normalizeTheme('  ')).toBe(DEFAULT_THEME);
		expect(normalizeTheme('hacker')).toBe(DEFAULT_THEME);
	});

	it('keeps every known theme id', () => {
		for (const th of THEMES) expect(normalizeTheme(th.id)).toBe(th.id);
	});
});
