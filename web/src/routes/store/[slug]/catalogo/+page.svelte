<script lang="ts">
	import type { PageData } from './$types';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import { t } from '$lib/i18n';
	import { invalidateAll } from '$app/navigation';
	import { SlidersHorizontal } from 'lucide-svelte';

	let { data }: { data: PageData } = $props();

	const storeName = $derived(data.store?.displayName ?? data.storeSlug);
	const base = $derived('/store/' + data.storeSlug + '/catalogo');

	// Categories derived from the store's own products (deduped, sorted) — a bookstore shows its genres,
	// not the demo's "footwear". The "all" pill stays translated; merchant categories show as typed.
	const categories = $derived([
		'',
		...Array.from(new Set(data.products.map((p) => p.category).filter(Boolean))).sort()
	]);
	const visibleProducts = $derived(
		data.category ? data.products.filter((p) => p.category === data.category) : data.products
	);
	const label = (c: string) => (c === '' ? $t('catalog.all') : c);
	const catHref = (c: string) => (c ? `${base}?category=${c}` : base);
</script>

<svelte:head>
	<title>{storeName} — {$t('catalog.heading')}</title>
</svelte:head>

<section class="mx-auto w-full max-w-container px-4 pb-24 pt-12 md:px-12 md:pt-16">
	<header class="mb-12 grid grid-cols-12 gap-px md:mb-16">
		<div class="col-span-12 md:col-span-8">
			<h1 class="font-sans uppercase leading-none tracking-tighter text-headline-lg md:text-display-xl">
				{storeName}
			</h1>
		</div>
		<div
			class="col-span-12 mt-6 flex flex-col justify-end pb-2 font-mono text-metadata-sm uppercase opacity-80 md:col-span-4 md:mt-auto"
		>
			<p>{$t('catalog.heading')}</p>
			<p>{$t('catalog.showing', { n: visibleProducts.length })}</p>
		</div>
	</header>

	<!-- Filter bar — only when the store has more than one category to filter by. -->
	{#if categories.length > 1}
		<div class="mb-10 flex items-center gap-6 border-y border-border py-3 font-mono text-metadata-sm uppercase">
			<span class="flex items-center gap-2 text-text-muted">
				<SlidersHorizontal size={14} />
				{$t('catalog.filter')}
			</span>
			<div class="flex flex-wrap gap-4">
				{#each categories as category (category)}
					<a
						href={catHref(category)}
						class="transition-colors hover:text-accent {data.category === category
							? 'text-text underline decoration-1 underline-offset-4'
							: 'text-text-muted'}"
					>
						{label(category)}
					</a>
				{/each}
			</div>
		</div>
	{/if}

	{#if data.failed}
		<div class="flex flex-col items-center gap-4 border border-border py-24 text-center">
			<p class="font-mono text-metadata-sm uppercase tracking-widest text-text-muted">
				{$t('catalog.errorLoad')}
			</p>
			<button onclick={() => invalidateAll()} class="btn-ghost">{$t('common.retry')}</button>
		</div>
	{:else if visibleProducts.length === 0}
		<div class="flex flex-col items-center gap-3 border border-border py-24 text-center">
			<h2 class="font-sans text-headline-md">{$t('catalog.emptyTitle')}</h2>
			<p class="font-mono text-metadata-sm uppercase tracking-widest text-text-muted">
				{$t('catalog.emptyBody')}
			</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
			{#each visibleProducts as product (product.id)}
				<ProductCard {product} base={'/store/' + data.storeSlug} />
			{/each}
		</div>
	{/if}
</section>
