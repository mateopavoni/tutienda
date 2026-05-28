import type { Locale } from '$lib/i18n';

// money renders integer cents as a localized currency string.
export function money(cents: number, currency = 'USD', locale: Locale = 'en'): string {
	return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(cents / 100);
}
