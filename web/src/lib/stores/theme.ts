import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'system' | 'light' | 'dark';
const STORAGE_KEY = 'theme';

function initialTheme(): Theme {
	if (!browser) return 'system';
	return (localStorage.getItem(STORAGE_KEY) as Theme) ?? 'system';
}

export const theme = writable<Theme>(initialTheme());

const THEME_ANIMS = ['circle', 'wipe-x', 'wipe-y'];

// setTheme applies the change through a View Transition (a random reveal animation,
// anchored to the click point) when the browser supports it; falls back to a plain set.
export function setTheme(value: Theme, e?: MouseEvent) {
	if (!browser || !document.startViewTransition || !e || (e.clientX === 0 && e.clientY === 0)) {
		theme.set(value);
		return;
	}
	document.documentElement.style.setProperty('--theme-x', `${e.clientX}px`);
	document.documentElement.style.setProperty('--theme-y', `${e.clientY}px`);
	document.documentElement.dataset.themeAnim = THEME_ANIMS[Math.floor(Math.random() * THEME_ANIMS.length)];
	document.startViewTransition(() => theme.set(value));
}

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
