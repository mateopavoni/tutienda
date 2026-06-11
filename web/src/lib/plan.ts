// Client-side mirror of the membership tiers in services/internal/platform/plan/plan.go. The Go package
// is the source of truth and the one that actually enforces (a 402 from the gateway is the real wall);
// this copy only drives the UI — which tools to dim, which upgrade to suggest. Keep the two in sync.

export type Tier = 'free' | 'pro' | 'scale';

export type Feature = 'drops' | 'customLogo' | 'analytics' | 'removeBranding';

export const UNLIMITED = -1;

interface Entitlement {
	label: string;
	price: string;
	features: Feature[];
	productLimit: number;
}

export const PLANS: Record<Tier, Entitlement> = {
	free: { label: 'Free', price: '$0', features: [], productLimit: 15 },
	pro: {
		label: 'Pro',
		price: '$29',
		features: ['drops', 'customLogo', 'analytics'],
		productLimit: 300
	},
	scale: {
		label: 'Scale',
		price: '$99',
		features: ['drops', 'customLogo', 'analytics', 'removeBranding'],
		productLimit: UNLIMITED
	}
};

// TIERS lists plans in upgrade order (cheapest first).
export const TIERS: Tier[] = ['free', 'pro', 'scale'];

/** normalizeTier coerces unknown/empty input to 'free', matching the Go Normalize (never silently unlock). */
export function normalizeTier(s: string | null | undefined): Tier {
	return s && s in PLANS ? (s as Tier) : 'free';
}

/** hasFeature reports whether a tier unlocks a feature. */
export function hasFeature(tier: string | null | undefined, f: Feature): boolean {
	return PLANS[normalizeTier(tier)].features.includes(f);
}

/** productLimit returns the tier's product cap, or UNLIMITED (-1). */
export function productLimit(tier: string | null | undefined): number {
	return PLANS[normalizeTier(tier)].productLimit;
}

/** cheapestTierWith returns the lowest tier that unlocks a feature (for an upgrade prompt). */
export function cheapestTierWith(f: Feature): Tier {
	return TIERS.find((t) => PLANS[t].features.includes(f)) ?? 'scale';
}
