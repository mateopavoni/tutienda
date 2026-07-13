<script lang="ts">
	import type { PageData } from './$types';
	import StorefrontSections from '$lib/components/StorefrontSections.svelte';
	import { enabledSections } from '$lib/layout';
	import { storeInfo } from '$lib/stores/storeInfo';

	let { data }: { data: PageData } = $props();

	// storeInfo starts as the SSR-resolved store and is refreshed by a background poll (see +layout.svelte)
	// so a merchant's saved changes show up here without the visitor reloading.
	const store = $derived($storeInfo ?? data.store);

	// Heading is the store's own name (multi-tenant), not the old "Dynamic Catalog" showcase title.
	const storeName = $derived(store?.displayName ?? data.storeSlug);
	const base = $derived('/store/' + data.storeSlug);
	const catalogHref = $derived(base + '/catalogo');

	// The merchant-arranged sections, in order. The grid ("catalog") is always present; the rest are the
	// optional blocks they enabled. Everything outside this list (nav/footer/tokens) is fixed across stores.
	const sections = $derived(enabledSections(store?.settings?.layout));

	// The home is hero-focused: the catalog block here is just a preview (the first few products) that links
	// to the dedicated /catalogo page, where the full filterable grid lives.
	const PREVIEW_COUNT = 6;
	const featured = $derived(data.products.slice(0, PREVIEW_COUNT));
</script>

<svelte:head>
	<title>{storeName} — {storeName}</title>
</svelte:head>

<StorefrontSections {sections} {featured} {base} {catalogHref} {storeName} />
