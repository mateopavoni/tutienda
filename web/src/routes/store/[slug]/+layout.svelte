<script lang="ts">
	import Nav from '$lib/components/Nav.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import CartDrawer from '$lib/components/CartDrawer.svelte';
	import { setBrowserSlug } from '$lib/tenant';
	import { loadCustomer } from '$lib/stores/customer';
	import { enabledPages } from '$lib/pages';
	import { normalizeTheme } from '$lib/theme';

	let { children, data } = $props();

	// Seed the browser API client with the resolved store slug and hydrate the buyer session for this
	// store; $effect keeps both in sync across navigations (and when switching between stores).
	$effect(() => {
		setBrowserSlug(data.storeSlug);
		loadCustomer(data.storeSlug);
	});

	// Per-store accent: the only chromatic token a store can override. Everything else stays fixed.
	// We guard against a near-white accent so the white-on-accent buttons never lose their label.
	let accentVar = $derived(safeAccent(data.store?.settings?.accentColor));

	// Per-store visual theme (fonts + palette); see web/src/lib/theme.ts + app.css [data-store-theme].
	// Defaults to monolith on any unknown/legacy value. The accent above still overrides the theme accent.
	let storeTheme = $derived(normalizeTheme(data.store?.settings?.theme));

	function safeAccent(hex: string | undefined): string | undefined {
		if (!hex) return undefined;
		const m = /^#?([0-9a-fA-F]{6})$/.exec(hex.trim());
		if (!m) return undefined;
		const n = parseInt(m[1], 16);
		const r = (n >> 16) & 255;
		const g = (n >> 8) & 255;
		const b = n & 255;
		// Relative luminance: if the accent is too light, white text on it would vanish — fall back to the
		// design-system accent rather than render an unreadable button (a real multi-tenant contrast bug).
		const lum = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
		if (lum > 0.7) return undefined;
		return `${r} ${g} ${b}`;
	}
</script>

<div
	class="flex min-h-screen flex-col bg-bg text-text"
	data-store-theme={storeTheme}
	style={accentVar ? `--accent: ${accentVar}` : undefined}
>
	<Nav store={data.store} slug={data.storeSlug} />
	<main class="flex-1 pt-16">
		{@render children()}
	</main>
	<Footer
		storeName={data.store?.displayName}
		slug={data.storeSlug}
		pages={enabledPages(data.store?.settings?.pages)}
	/>
	<CartDrawer />
</div>
