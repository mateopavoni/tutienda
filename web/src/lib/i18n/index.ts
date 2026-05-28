import { writable, derived, type Readable } from 'svelte/store';
import { browser } from '$app/environment';
import { en, type Messages } from './en';
import { es } from './es';
import { ptBR } from './pt-br';

export type Locale = 'en' | 'es' | 'pt-BR';

export const locales: { code: Locale; label: string }[] = [
	{ code: 'en', label: 'EN' },
	{ code: 'es', label: 'ES' },
	{ code: 'pt-BR', label: 'PT' }
];

const dictionaries: Record<Locale, Messages> = { en, es, 'pt-BR': ptBR };
const STORAGE_KEY = 'locale';

function initialLocale(): Locale {
	if (!browser) return 'en';
	const saved = localStorage.getItem(STORAGE_KEY) as Locale | null;
	if (saved && saved in dictionaries) return saved;
	const nav = navigator.language.toLowerCase();
	if (nav.startsWith('es')) return 'es';
	if (nav.startsWith('pt')) return 'pt-BR';
	return 'en';
}

export const locale = writable<Locale>(initialLocale());
if (browser) {
	locale.subscribe((value) => localStorage.setItem(STORAGE_KEY, value));
}

// resolve walks a dotted key ("catalog.title") through the dictionary.
function resolve(messages: Messages, key: string): string {
	const parts = key.split('.');
	let current: unknown = messages;
	for (const part of parts) {
		if (current && typeof current === 'object' && part in current) {
			current = (current as Record<string, unknown>)[part];
		} else {
			return key; // missing key surfaces visibly instead of crashing
		}
	}
	return typeof current === 'string' ? current : key;
}

function interpolate(text: string, params?: Record<string, string | number>): string {
	if (!params) return text;
	return text.replace(/\{(\w+)\}/g, (_, name) => (name in params ? String(params[name]) : `{${name}}`));
}

export type Translator = (key: string, params?: Record<string, string | number>) => string;

// t is a reactive translator: it re-derives whenever the locale changes.
export const t: Readable<Translator> = derived(locale, ($locale) => {
	const messages = dictionaries[$locale];
	return (key: string, params?: Record<string, string | number>) =>
		interpolate(resolve(messages, key), params);
});
