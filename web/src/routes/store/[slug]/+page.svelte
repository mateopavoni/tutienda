<script lang="ts">
	import type { PageData } from './$types';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import { t } from '$lib/i18n';
	import { invalidateAll } from '$app/navigation';
	import { SlidersHorizontal, ArrowRight } from 'lucide-svelte';
	import { enabledSections } from '$lib/layout';

	let { data }: { data: PageData } = $props();

	// Heading is the store's own name (multi-tenant), not the old "Dynamic Catalog" showcase title.
	const storeName = $derived(data.store?.displayName ?? data.storeSlug);
	const base = $derived('/store/' + data.storeSlug);

	// The merchant-arranged sections, in order. The grid ("catalog") is always present; the rest are the
	// optional blocks they enabled. Everything outside this list (nav/footer/tokens) is fixed across stores.
	const sections = $derived(enabledSections(data.store?.settings?.layout));

	// Categories are derived from the store's own products (deduped, sorted) so the filter fits any kind
	// of store — a bookstore shows its genres, not the demo's "footwear/outerwear". The grid filters in
	// the page; the "all" pill stays translated, while merchant categories show their own label as typed.
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

{#each sections as section (section.type)}
	{#if section.type === 'announcement'}
		<div class="border-b border-border bg-inverse px-4 py-2 text-center font-mono text-metadata-sm uppercase tracking-[0.1em] text-on-inverse">
			{section.body}
		</div>
	{:else if section.type === 'hero'}
		<section class="relative border-b border-border">
			{#if section.imageUrl}
				<img src={section.imageUrl} alt={section.heading ?? storeName} class="h-[55vh] w-full object-cover" />
				<div class="absolute inset-0 bg-black/30"></div>
			{/if}
			<div
				class="mx-auto flex w-full max-w-container flex-col items-start gap-6 px-4 py-20 md:px-12 md:py-28
					{section.imageUrl ? 'absolute inset-0 justify-center text-on-inverse' : ''}"
			>
				{#if section.heading}
					<h2 class="max-w-3xl font-sans uppercase leading-none tracking-tighter text-display-xl">
						{section.heading}
					</h2>
				{/if}
				{#if section.body}
					<p class="max-w-xl font-sans text-body-lg {section.imageUrl ? '' : 'text-text-muted'}">{section.body}</p>
				{/if}
				{#if section.ctaLabel}
					<a href={base} class="flex items-center gap-3 btn-accent">
						{section.ctaLabel}
						<ArrowRight size={16} />
					</a>
				{/if}
			</div>
		</section>
	{:else if section.type === 'feature'}
		<section class="border-b border-border">
			<div class="mx-auto grid w-full max-w-container grid-cols-1 items-stretch md:grid-cols-2">
				{#if section.imageUrl}
					<img src={section.imageUrl} alt={section.heading ?? ''} class="h-full min-h-[20rem] w-full object-cover" />
				{/if}
				<div class="flex flex-col items-start justify-center gap-5 px-4 py-16 md:px-12">
					{#if section.heading}
						<h2 class="font-sans uppercase leading-none tracking-tighter text-headline-lg">{section.heading}</h2>
					{/if}
					{#if section.body}
						<p class="max-w-md whitespace-pre-line font-sans text-body-lg leading-relaxed text-text-muted">{section.body}</p>
					{/if}
					{#if section.ctaLabel}
						<a href={base} class="flex items-center gap-3 btn-accent">
							{section.ctaLabel}
							<ArrowRight size={16} />
						</a>
					{/if}
				</div>
			</div>
		</section>
	{:else if section.type === 'about'}
		<section class="border-b border-border">
			<div class="mx-auto grid w-full max-w-container grid-cols-1 gap-8 px-4 py-20 md:grid-cols-12 md:px-12">
				<h2 class="font-sans uppercase leading-none tracking-tighter text-headline-lg md:col-span-4">
					{section.heading}
				</h2>
				<p class="whitespace-pre-line font-sans text-body-lg leading-relaxed text-text-muted md:col-span-8">
					{section.body}
				</p>
			</div>
		</section>
	{:else if section.type === 'contact'}
		<section class="border-b border-border bg-surface">
			<div class="mx-auto grid w-full max-w-container grid-cols-1 gap-8 px-4 py-16 md:grid-cols-12 md:px-12">
				<h2 class="font-sans uppercase leading-none tracking-tighter text-headline-md md:col-span-4">
					{section.heading}
				</h2>
				<p class="whitespace-pre-line font-mono text-body-md leading-relaxed text-text-muted md:col-span-8">
					{section.body}
				</p>
			</div>
		</section>
	{:else if section.type === 'gallery'}
		{#if section.items?.some((it) => it.imageUrl)}
			<section class="border-b border-border">
				<div class="mx-auto w-full max-w-container px-4 py-16 md:px-12">
					{#if section.heading}
						<h2 class="mb-8 font-sans uppercase leading-none tracking-tighter text-headline-md">{section.heading}</h2>
					{/if}
					<div class="grid grid-cols-2 gap-2 md:grid-cols-3">
						{#each section.items ?? [] as item (item.imageUrl)}
							{#if item.imageUrl}
								<figure class="relative">
									<img src={item.imageUrl} alt={item.heading ?? ''} class="aspect-square w-full object-cover" />
									{#if item.heading}
										<figcaption class="absolute bottom-0 left-0 right-0 bg-black/50 px-3 py-1 font-mono text-metadata-sm text-on-inverse">
											{item.heading}
										</figcaption>
									{/if}
								</figure>
							{/if}
						{/each}
					</div>
				</div>
			</section>
		{/if}
	{:else if section.type === 'faq'}
		{#if section.items?.some((it) => it.heading || it.body)}
			<section class="border-b border-border">
				<div class="mx-auto grid w-full max-w-container grid-cols-1 gap-8 px-4 py-16 md:grid-cols-12 md:px-12">
					<h2 class="font-sans uppercase leading-none tracking-tighter text-headline-md md:col-span-4">
						{section.heading}
					</h2>
					<dl class="flex flex-col divide-y divide-border md:col-span-8">
						{#each section.items ?? [] as item (item.heading)}
							{#if item.heading || item.body}
								<div class="py-4">
									<dt class="font-sans text-body-lg">{item.heading}</dt>
									{#if item.body}
										<dd class="mt-2 whitespace-pre-line font-sans text-body-md leading-relaxed text-text-muted">{item.body}</dd>
									{/if}
								</div>
							{/if}
						{/each}
					</dl>
				</div>
			</section>
		{/if}
	{:else if section.type === 'catalog'}
		<section class="mx-auto w-full max-w-container px-4 pb-24 pt-8 md:px-12">
			<header class="mb-12 grid grid-cols-12 gap-px md:mb-20">
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
						<ProductCard {product} {base} />
					{/each}
				</div>
			{/if}
		</section>
	{/if}
{/each}
