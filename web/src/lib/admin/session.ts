// Admin session: two tokens live in localStorage. The merchant token (from login/signup) authorizes
// accounts calls — listing/creating stores and minting a store token. The store token is scoped to one
// store and authorizes the admin catalog/inventory writes. Keeping them in localStorage keeps the admin
// a pure client-side app (ssr is disabled for /admin).
import { browser } from '$app/environment';

const MERCHANT_KEY = 'admin.merchantToken';
const STORE_KEY = 'admin.storeToken';
const ACTIVE_KEY = 'admin.activeStore';
const ROLE_KEY = 'admin.role';

export interface ActiveStore {
	id: string;
	slug: string;
	displayName: string;
}

function get(key: string): string {
	return browser ? (localStorage.getItem(key) ?? '') : '';
}
function set(key: string, value: string): void {
	if (browser) localStorage.setItem(key, value);
}
function del(key: string): void {
	if (browser) localStorage.removeItem(key);
}

export const session = {
	merchantToken: () => get(MERCHANT_KEY),
	storeToken: () => get(STORE_KEY),
	activeStore(): ActiveStore | null {
		const raw = get(ACTIVE_KEY);
		return raw ? (JSON.parse(raw) as ActiveStore) : null;
	},
	setMerchantToken: (t: string) => set(MERCHANT_KEY, t),
	role: () => get(ROLE_KEY),
	isAdmin: () => get(ROLE_KEY) === 'admin',
	setRole: (r: string) => set(ROLE_KEY, r),
	setStore(token: string, store: ActiveStore) {
		set(STORE_KEY, token);
		set(ACTIVE_KEY, JSON.stringify(store));
	},
	clearStore() {
		del(STORE_KEY);
		del(ACTIVE_KEY);
	},
	clearAll() {
		del(MERCHANT_KEY);
		del(STORE_KEY);
		del(ACTIVE_KEY);
		del(ROLE_KEY);
	}
};
