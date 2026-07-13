<script lang="ts">
	import type { Product } from '$lib/types';
	import type { Section } from '$lib/layout';
	import { HERO_OVERLAY_DEFAULT } from '$lib/layout';
	import ProductCard from '$lib/components/ProductCard.svelte';
	import { t } from '$lib/i18n';
	import { ArrowRight, ChevronDown } from 'lucide-svelte';

	// Renders a store's enabled sections in order, around the fixed product grid. Shared by the real
	// storefront (store/[slug]/+page.svelte) and the live preview in the /app layout editor, so a
	// merchant's unsaved edits render with the exact markup visitors will see.
	let {
		sections,
		featured,
		base,
		catalogHref,
		storeName
	}: { sections: Section[]; featured: Product[]; base: string; catalogHref: string; storeName: string } =
		$props();
</script>

{#each sections as section (section.type)}
	{#if section.type === 'announcement'}
		<div class="border-b border-border bg-inverse px-4 py-2 text-center font-mono text-metadata-sm uppercase tracking-[0.1em] text-on-inverse">
			{section.body}
		</div>
	{:else if section.type === 'hero'}
		<!-- Overlay is a left-to-right gradient, not a flat tint: strongest where the (left-aligned) text
		     sits, fading to transparent so the rest of the photo shows through untouched — the standard
		     photo-hero effect. The tint follows the visitor's light/dark mode (white veil + dark text in
		     light mode, black veil + light text in dark mode) via the `dark:` variant, not a merchant
		     choice — text color is left to inherit the page's own text-text, which is already the right
		     shade for each mode, so no JS branching is needed either. -->
		{@const overlayAlpha = (section.overlay ?? HERO_OVERLAY_DEFAULT) / 100}
		<section class="relative flex min-h-screen w-full items-center border-b border-border">
			{#if section.imageUrl}
				<img src={section.imageUrl} alt={section.heading ?? storeName} class="absolute inset-0 h-full w-full object-cover" />
				<div
					class="absolute inset-0 dark:hidden"
					style="background: linear-gradient(to right, rgba(255,255,255,{overlayAlpha}) 0%, rgba(255,255,255,0) 65%)"
				></div>
				<div
					class="absolute inset-0 hidden dark:block"
					style="background: linear-gradient(to right, rgba(0,0,0,{overlayAlpha}) 0%, rgba(0,0,0,0) 65%)"
				></div>
			{/if}
			<div class="relative mx-auto flex w-full max-w-container flex-col items-start gap-6 px-4 py-20 md:px-12 md:py-28">
				{#if section.heading}
					<h2 class="max-w-3xl font-sans uppercase leading-none tracking-tighter text-display-xl">
						{section.heading}
					</h2>
				{/if}
				{#if section.body}
					<p class="max-w-xl font-sans text-body-lg {section.imageUrl ? '' : 'text-text-muted'}">{section.body}</p>
				{/if}
				{#if section.ctaLabel}
					<a href={catalogHref} class="flex items-center gap-3 btn-accent">
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
						<a href={catalogHref} class="flex items-center gap-3 btn-accent">
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
					<div class="grid grid-cols-2 gap-3 md:grid-cols-3">
						{#each section.items ?? [] as item (item.imageUrl)}
							{#if item.imageUrl}
								<figure class="group relative overflow-hidden brutal-border">
									<img
										src={item.imageUrl}
										alt={item.heading ?? ''}
										class="aspect-square w-full object-cover transition-transform duration-300 group-hover:scale-105"
									/>
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
					<div class="flex flex-col divide-y divide-border md:col-span-8">
						{#each section.items ?? [] as item (item.heading)}
							{#if item.heading || item.body}
								<details class="group py-4">
									<summary class="flex cursor-pointer list-none items-center justify-between gap-4 font-sans text-body-lg marker:content-none [&::-webkit-details-marker]:hidden">
										{item.heading}
										<ChevronDown size={18} class="shrink-0 transition-transform group-open:rotate-180" />
									</summary>
									{#if item.body}
										<p class="mt-2 whitespace-pre-line font-sans text-body-md leading-relaxed text-text-muted">{item.body}</p>
									{/if}
								</details>
							{/if}
						{/each}
					</div>
				</div>
			</section>
		{/if}
	{:else if section.type === 'catalog'}
		<!-- Home preview: a taste of the catalog (first few products) linking to the full /catalogo page. -->
		{#if featured.length > 0}
			<section class="mx-auto w-full max-w-container px-4 pb-24 pt-12 md:px-12 md:pt-16">
				<header class="mb-10 flex items-end justify-between gap-6 border-b border-border pb-4">
					<h2 class="font-sans uppercase leading-none tracking-tighter text-headline-md md:text-headline-lg">
						{$t('catalog.featured')}
					</h2>
					<a
						href={catalogHref}
						class="flex shrink-0 items-center gap-2 font-mono text-metadata-sm uppercase tracking-[0.1em] text-accent transition-colors hover:text-text"
					>
						{$t('catalog.viewAll')}
						<ArrowRight size={14} />
					</a>
				</header>
				<div class="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
					{#each featured as product (product.id)}
						<ProductCard {product} {base} />
					{/each}
				</div>
			</section>
		{/if}
	{/if}
{/each}
