import { describe, it, expect } from 'vitest';
import { normalizeTier, hasFeature, productLimit, cheapestTierWith, UNLIMITED } from './plan';

// This mirrors services/internal/platform/plan/plan.go. If they drift, the UI dims the wrong tools.
describe('plan mirror (matches Go entitlements)', () => {
	it('normalizes unknown/empty/tampered tiers to free (never silently unlocks)', () => {
		expect(normalizeTier(undefined)).toBe('free');
		expect(normalizeTier('')).toBe('free');
		expect(normalizeTier('enterprise')).toBe('free');
		expect(normalizeTier('pro')).toBe('pro');
	});

	it('gates features by tier', () => {
		expect(hasFeature('free', 'drops')).toBe(false);
		expect(hasFeature('pro', 'drops')).toBe(true);
		expect(hasFeature('pro', 'removeBranding')).toBe(false); // scale-only
		expect(hasFeature('scale', 'removeBranding')).toBe(true);
		expect(hasFeature('bogus', 'drops')).toBe(false); // tampered → free
	});

	it('reports product limits, with scale unlimited', () => {
		expect(productLimit('free')).toBe(15);
		expect(productLimit('pro')).toBe(300);
		expect(productLimit('scale')).toBe(UNLIMITED);
	});

	it('finds the cheapest tier that unlocks a feature', () => {
		expect(cheapestTierWith('drops')).toBe('pro');
		expect(cheapestTierWith('removeBranding')).toBe('scale');
	});
});
