<script lang="ts">
	import Nav from '$lib/components/Nav.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import CartDrawer from '$lib/components/CartDrawer.svelte';
	import { setBrowserSlug } from '$lib/tenant';
	import { loadCustomer } from '$lib/stores/customer';
	import { cart } from '$lib/stores/cart';
	import { enabledPages } from '$lib/pages';
	import { normalizeTheme } from '$lib/theme';
	import { theme } from '$lib/stores/theme';
	import { storeInfo, fetchStoreInfo } from '$lib/stores/storeInfo';
	import { safeAccent, safeTitleColor } from '$lib/colorGuard';
	import { browser } from '$app/environment';

	let { children, data } = $props();

	// Seed the browser API client with the resolved store slug and hydrate the buyer session for this
	// store; $effect keeps both in sync across navigations (and when switching between stores).
	$effect(() => {
		setBrowserSlug(data.storeSlug);
		loadCustomer(data.storeSlug);
		cart.hydrate(data.storeSlug);
	});

	// Live storefront settings: reset to the SSR snapshot on navigation, then poll so a merchant's saved
	// branding/layout edits reach an already-open tab without a manual reload (same pattern as the live
	// stock counter on the product page — poll while relevant, cheap enough at this interval).
	$effect(() => {
		storeInfo.set(data.store);
		const slug = data.storeSlug;
		const id = setInterval(async () => {
			const fresh = await fetchStoreInfo(slug);
			if (fresh) storeInfo.set(fresh);
		}, 20000);
		return () => clearInterval(id);
	});

	// Resolved light/dark scheme (mirrors the logic in $lib/stores/theme's applyTheme) so the accent can
	// be adjusted for the mode actually on screen, not just the light-mode case.
	let prefersDark = $state(false);
	$effect(() => {
		if (!browser) return;
		const mq = window.matchMedia('(prefers-color-scheme: dark)');
		const update = () => (prefersDark = mq.matches);
		update();
		mq.addEventListener('change', update);
		return () => mq.removeEventListener('change', update);
	});
	let isDark = $derived($theme === 'dark' || ($theme === 'system' && prefersDark));

	// $storeInfo is the live mirror seeded above (falls back to the SSR data before the effect runs).
	const store = $derived($storeInfo ?? data.store);

	// Per-store accent: the only chromatic token a store can override. Everything else stays fixed.
	let accentVar = $derived(safeAccent(store?.settings?.accentColor, isDark));
	// Per-store title color: applied to each page's main headline (see the `--title-color` consumers in
	// StorefrontSections.svelte, catalogo/product/[page] headings) — falls back to the design system's
	// text-text when unset or when it'd be unreadable against the page background.
	let titleColorVar = $derived(safeTitleColor(store?.settings?.titleColor, isDark));

	// Per-store visual theme (fonts + palette); see web/src/lib/theme.ts + app.css [data-store-theme].
	// Defaults to monolith on any unknown/legacy value. The accent above still overrides the theme accent.
	let storeTheme = $derived(normalizeTheme(store?.settings?.theme));
</script>

<div
	class="flex min-h-screen flex-col bg-bg text-text"
	data-store-theme={storeTheme}
	style="{accentVar ? `--accent: ${accentVar};` : ''}{titleColorVar ? `--title-color: ${titleColorVar};` : ''}"
>
	<Nav {store} slug={data.storeSlug} />
	<main class="flex-1 pt-16">
		{@render children()}
	</main>
	<Footer storeName={store?.displayName} slug={data.storeSlug} pages={enabledPages(store?.settings?.pages)} />
	<CartDrawer slug={data.storeSlug} />
</div>
