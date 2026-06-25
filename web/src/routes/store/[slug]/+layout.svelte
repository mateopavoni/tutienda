<script lang="ts">
	import Nav from '$lib/components/Nav.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import CartDrawer from '$lib/components/CartDrawer.svelte';
	import { setBrowserSlug } from '$lib/tenant';
	import { enabledPages } from '$lib/pages';

	let { children, data } = $props();

	// Seed the browser API client with the resolved store slug; $effect keeps it in sync across navigations.
	$effect(() => {
		setBrowserSlug(data.storeSlug);
	});

	// Per-store accent: the only chromatic token a store can override. Everything else stays fixed.
	// We guard against a near-white accent so the white-on-accent buttons never lose their label.
	let accentVar = $derived(safeAccent(data.store?.settings?.accentColor));

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

<div class="flex min-h-screen flex-col" style={accentVar ? `--accent: ${accentVar}` : undefined}>
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
