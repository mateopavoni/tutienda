import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'system' | 'light' | 'dark';
const STORAGE_KEY = 'theme';

function initialTheme(): Theme {
	if (!browser) return 'system';
	return (localStorage.getItem(STORAGE_KEY) as Theme) ?? 'system';
}

export const theme = writable<Theme>(initialTheme());

// applyTheme resolves "system" against the OS preference and toggles the root class.
export function applyTheme(value: Theme) {
	if (!browser) return;
	const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
	const dark = value === 'dark' || (value === 'system' && prefersDark);
	document.documentElement.classList.toggle('dark', dark);
	document.documentElement.classList.toggle('light', !dark);
	localStorage.setItem(STORAGE_KEY, value);
}

if (browser) {
	theme.subscribe(applyTheme);
}
