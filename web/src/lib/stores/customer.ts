// Storefront customer session. A buyer token is bound to ONE store (the gateway signs the tenant into
// it), so it's persisted per slug under `customer:<slug>`. One active store per tab, so the token is also
// held in a module-local for the checkout/orders calls. Separate from the merchant session ($lib/admin).
import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface CustomerProfile {
	id: string;
	email: string;
	name?: string;
}

interface Persisted {
	token: string;
	profile: CustomerProfile;
}

const key = (slug: string) => 'customer:' + slug;

// Reactive profile for the chrome (Nav account link, account page). null = not signed in here.
export const customer = writable<CustomerProfile | null>(null);

let activeSlug = '';
let activeToken = '';

// customerToken returns the bearer for the active store, or '' for a guest.
export function customerToken(): string {
	return activeToken;
}

// loadCustomer hydrates the session from localStorage for the given store (called on storefront mount).
export function loadCustomer(slug: string): void {
	activeSlug = slug;
	activeToken = '';
	if (!browser) {
		customer.set(null);
		return;
	}
	try {
		const raw = localStorage.getItem(key(slug));
		if (raw) {
			const p = JSON.parse(raw) as Persisted;
			activeToken = p.token;
			customer.set(p.profile);
			return;
		}
	} catch {
		// fall through to signed-out
	}
	customer.set(null);
}

// setCustomerSession persists a fresh login/signup for the store and makes it active.
export function setCustomerSession(slug: string, token: string, profile: CustomerProfile): void {
	activeSlug = slug;
	activeToken = token;
	customer.set(profile);
	if (browser) localStorage.setItem(key(slug), JSON.stringify({ token, profile }));
}

// clearCustomer signs the buyer out of the store.
export function clearCustomer(slug: string): void {
	if (slug === activeSlug) activeToken = '';
	customer.set(null);
	if (browser) localStorage.removeItem(key(slug));
}
