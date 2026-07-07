<script lang="ts">
	import type { PageData } from './$types';
	import { t, locale } from '$lib/i18n';
	import { money } from '$lib/format';
	import { cart } from '$lib/stores/cart';
	import { cartOpen } from '$lib/stores/ui';
	import { dropState } from '$lib/drop';
	import { fetchStock } from '$lib/api/client';
	import DropCountdown from '$lib/components/DropCountdown.svelte';
	import { ArrowRight, ArrowLeft, Radio } from 'lucide-svelte';

	let { data }: { data: PageData } = $props();
	const product = $derived(data.product);
	const storeName = $derived(data.store?.displayName ?? data.storeSlug);
	const backHref = $derived('/store/' + data.storeSlug);
	const skus = $derived(product.variants.map((v) => v.sku));

	// Image gallery: the product's images, falling back to the single primary for legacy/seed products.
	const gallery = $derived(
		product.images && product.images.length
			? product.images
			: product.imageUrl
				? [product.imageUrl]
				: []
	);
	let activeImage = $state(0);
	// Reset to the first image whenever the product changes (the component is reused across navigations).
	$effect(() => {
		void product.id;
		activeImage = 0;
	});
	const mainImage = $derived(gallery[activeImage] ?? product.imageUrl);

	// A once-per-second clock drives the drop state, so an upcoming drop flips to live on its own (and the
	// state stays correct when SvelteKit reuses this component across product navigations).
	let now = $state(Date.now());
	$effect(() => {
		const id = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(id);
	});
	const phase = $derived(dropState(product.dropAt, now));
	const isUpcoming = $derived(phase === 'upcoming');
	const isLiveDrop = $derived(product.dropAt != null && phase === 'live');

	// Live stock: reset to the server-rendered snapshot whenever the product changes, then refreshed by
	// polling while a drop is live so the counter ticks down in near real time as buyers take the last units.
	let live = $state<Record<string, number>>({});
	$effect(() => {
		live = { ...data.stockMap };
	});
	const liveTotal = $derived(product.variants.reduce((sum, v) => sum + (live[v.sku] ?? 0), 0));

	$effect(() => {
		if (!isLiveDrop) return;
		let active = true;
		const poll = async () => {
			try {
				const stocks = await fetchStock(skus);
				if (!active) return;
				const m: Record<string, number> = {};
				for (const s of stocks) m[s.sku] = s.available;
				live = m;
			} catch {
				/* transient: keep the last known counts */
			}
		};
		poll();
		const id = setInterval(poll, 4000);
		return () => {
			active = false;
			clearInterval(id);
		};
	});

	let selectedSku = $state<string | null>(null);
	const selectedVariant = $derived(product.variants.find((v) => v.sku === selectedSku) ?? null);
	const selectedAvailable = $derived(selectedSku ? (live[selectedSku] ?? 0) : 0);
	const canAdd = $derived(!isUpcoming && selectedVariant !== null && selectedAvailable > 0);

	function addToBag() {
		if (!selectedVariant) return;
		cart.add({
			sku: selectedVariant.sku,
			productId: product.id,
			name: product.name,
			label: selectedVariant.label,
			priceCents: product.priceCents,
			currency: product.currency,
			imageUrl: product.imageUrl
		});
		cartOpen.set(true);
	}
</script>

<svelte:head>
	<title>{storeName} — {product.name}</title>
</svelte:head>

<section class="flex flex-col lg:h-[calc(100vh-4rem)] lg:flex-row">
	<!-- Image gallery -->
	<div class="relative h-[60vh] w-full overflow-hidden bg-surface-variant lg:h-full lg:w-[70%]">
		<img src={mainImage} alt={product.name} class="h-full w-full object-cover" />
		{#if product.dropAt}
			<span
				class="absolute left-4 top-4 flex items-center gap-2 bg-accent px-3 py-1.5 font-mono text-metadata-sm uppercase tracking-[0.1em] text-on-accent"
			>
				{#if isUpcoming}
					{$t('drop.badge')}
				{:else}
					<Radio size={13} strokeWidth={2} /> {$t('drop.liveNow')}
				{/if}
			</span>
		{/if}
		{#if gallery.length > 1}
			<div class="absolute bottom-4 left-4 right-4 flex flex-wrap gap-2">
				{#each gallery as img, i (img)}
					<button
						type="button"
						onclick={() => (activeImage = i)}
						aria-label="View image {i + 1}"
						aria-pressed={activeImage === i}
						class="h-14 w-14 overflow-hidden border-2 bg-surface transition-opacity {activeImage === i
							? 'border-accent'
							: 'border-transparent opacity-70 hover:opacity-100'}"
					>
						<img src={img} alt="" class="h-full w-full object-cover" />
					</button>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Details -->
	<div class="flex w-full flex-col border-border lg:w-[30%] lg:overflow-y-auto lg:border-l">
		<div class="flex flex-1 flex-col gap-10 p-8 lg:p-12">
			<div
				class="flex items-center justify-between border-b border-border pb-4 font-mono text-metadata-sm uppercase text-text-muted"
			>
				<a href={backHref} class="flex items-center gap-2 transition-colors hover:text-text">
					<ArrowLeft size={14} />
					{$t('product.back')}
				</a>
				<span>{product.category}</span>
			</div>

			<div class="flex flex-col gap-4">
				<h1
					class="font-sans uppercase leading-none tracking-tighter text-headline-md md:text-headline-lg text-text"
				>
					{product.name}
				</h1>
				<div class="font-sans text-headline-md text-text">
					{money(product.priceCents, product.currency, $locale)}
				</div>
			</div>

			<p class="max-w-sm font-sans text-body-md leading-relaxed text-text-muted">
				{product.description}
			</p>

			{#if isUpcoming && product.dropAt}
				<!-- Upcoming drop: countdown, no purchase. Server-side gate backs this up. -->
				<div class="flex flex-col gap-4 border border-accent p-6">
					<span class="font-mono text-metadata-sm uppercase tracking-[0.1em] text-accent">
						{$t('drop.dropsIn')}
					</span>
					<div class="text-headline-lg text-text">
						<DropCountdown dropAt={product.dropAt} />
					</div>
					<p class="font-mono text-metadata-sm leading-relaxed text-text-muted">
						{$t('product.reserveNote')}
					</p>
				</div>
			{:else}
				{#if isLiveDrop}
					<!-- Live drop: real-time remaining counter on top of the no-oversell engine. -->
					<div
						class="flex items-center justify-between border border-accent px-4 py-3 font-mono text-metadata-sm uppercase tracking-[0.1em]"
					>
						<span class="flex items-center gap-2 text-accent">
							<Radio size={14} strokeWidth={2} />
							{$t('drop.live')}
						</span>
						<span class="tabular-nums text-text">{$t('drop.remaining', { n: liveTotal })}</span>
					</div>
				{/if}

				<!-- Size selector -->
				<div class="flex flex-col gap-4">
					<span class="font-mono text-metadata-sm uppercase text-text-muted">{$t('product.size')}</span>
					<div class="grid grid-cols-4 gap-px border border-border bg-border">
						{#each product.variants as variant (variant.sku)}
							{@const available = live[variant.sku] ?? 0}
							<button
								onclick={() => (selectedSku = variant.sku)}
								disabled={available <= 0}
								class="relative py-3 font-mono text-metadata-sm transition-colors
									{selectedSku === variant.sku ? 'bg-primary text-on-primary' : 'bg-surface text-text hover:bg-surface-variant'}
									{available <= 0 ? 'cursor-not-allowed opacity-40' : ''}"
								aria-pressed={selectedSku === variant.sku}
							>
								{variant.label}
								{#if available <= 0}
									<span
										class="pointer-events-none absolute inset-0 flex items-center justify-center"
										aria-hidden="true"
									>
										<span class="h-px w-[140%] -rotate-45 bg-text"></span>
									</span>
								{/if}
							</button>
						{/each}
					</div>
					{#if selectedVariant && selectedAvailable > 0 && selectedAvailable <= 10}
						<span class="font-mono text-metadata-sm uppercase text-accent">
							{$t('common.lowStock', { n: selectedAvailable })}
						</span>
					{/if}
				</div>

				<div class="mt-auto flex flex-col gap-4">
					<button
						onclick={addToBag}
						disabled={!canAdd}
						class="flex w-full items-center justify-center gap-3 btn-accent disabled:cursor-not-allowed disabled:opacity-40"
					>
						{canAdd ? $t('common.addToBag') : $t('product.selectSize')}
						<ArrowRight size={16} />
					</button>
					<p class="font-mono text-metadata-sm leading-relaxed text-text-muted">
						{$t('product.reserveNote')}
					</p>
				</div>
			{/if}
		</div>
	</div>
</section>
