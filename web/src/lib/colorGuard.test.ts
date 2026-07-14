import { describe, it, expect } from 'vitest';
import { safeAccent, safeTitleColor } from './colorGuard';

describe('colorGuard', () => {
	it('safeAccent rejects a too-light color and lightens a too-dark one in dark mode', () => {
		expect(safeAccent(undefined, false)).toBeUndefined();
		expect(safeAccent('#fefefe', false)).toBeUndefined(); // near-white -> unreadable button text
		expect(safeAccent('#5433eb', false)).toBe('84 51 235'); // mid-range -> passes through as-is
		expect(safeAccent('#1a1a1a', true)).not.toBe('26 26 26'); // near-black lightened for dark-mode text
	});

	it('safeTitleColor rejects colors that would vanish against the page background', () => {
		expect(safeTitleColor('#fefefe', false)).toBeUndefined(); // near-white text on light page
		expect(safeTitleColor('#0a0a0a', true)).toBeUndefined(); // near-black text on dark page
		expect(safeTitleColor('#5433eb', false)).toBe('84 51 235');
		expect(safeTitleColor('#5433eb', true)).toBe('84 51 235');
		expect(safeTitleColor('not-a-color', false)).toBeUndefined();
	});
});
