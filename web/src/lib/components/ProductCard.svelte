<script lang="ts">
	import type { Product } from '$lib/types';
	import { t, locale } from '$lib/i18n';
	import { money } from '$lib/format';
	import { dropState } from '$lib/drop';

	// `base` is the store path (/store/[slug]) so product links stay scoped to the active store.
	let { product, base }: { product: Product; base: string } = $props();
	// Static classification (no per-card ticking): the live countdown lives on the detail page.
	const state = $derived(dropState(product.dropAt));
</script>

<article class="group">
	<a href="{base}/product/{product.id}" class="block">
		<div class="relative w-full overflow-hidden border border-border bg-surface-variant">
			{#if state === 'upcoming'}
				<span
					class="absolute left-0 top-0 z-10 bg-accent px-2 py-1 font-mono text-metadata-sm uppercase tracking-[0.1em] text-on-accent"
				>
					{$t('drop.badge')}
				</span>
			{/if}
			<img
				src={product.imageUrl}
				alt={product.name}
				loading="lazy"
				class="aspect-[4/5] w-full object-cover grayscale transition-all duration-500 group-hover:grayscale-0"
			/>
		</div>
		<div class="mt-3 flex items-start justify-between border-t border-border pt-2 font-mono text-metadata-sm">
			<div>
				<h2 class="font-bold uppercase text-text">{product.name}</h2>
				<p class="text-text-muted">{$t('catalog.' + product.category)}</p>
			</div>
			<span class="font-bold text-text">{money(product.priceCents, product.currency, $locale)}</span>
		</div>
	</a>
</article>
