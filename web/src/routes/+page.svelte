<script lang="ts">
	// Marketing landing for the SaaS. This is NOT a storefront — it sells the platform. The actual stores
	// live under /store/[slug]. Brand name comes from $lib/brand.
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { ArrowRight } from 'lucide-svelte';

	// Editorial wayfinding: features are numbered, not iconed — numerals read more like a magazine index
	// than the product-y lucide glyphs, which fits the Monolith Editorial system.
	const features = ['feat1', 'feat2', 'feat3', 'feat4'];

	const plans = ['free', 'pro', 'scale'] as const;

	// Sample stores seeded by the backend (tenant.DemoStores) — surfaced here so prospective merchants
	// see the range of looks one platform produces. Names/taglines are store content, not UI copy.
	const examples = [
		{ slug: 'system-archive', name: 'SYSTEM ARCHIVE', theme: 'Monolith', blurb: 'Engineered garments, brutalist editorial.' },
		{ slug: 'casa-bruta', name: 'CASA BRUTA', theme: 'Boutique', blurb: 'Objects for the brutalist home.' },
		{ slug: 'papel-y-tinta', name: 'PAPEL & TINTA', theme: 'Terminal', blurb: 'An editorial bookshop on paper.' },
		{ slug: 'cafe-noventa', name: 'CAFÉ NOVENTA', theme: 'Pop', blurb: 'Single-origin coffee, severe standard.' }
	];
</script>

<svelte:head>
	<title>{BRAND} — {$t('landing.tagline')}</title>
	<meta name="description" content={$t('landing.subtitle')} />
	<meta property="og:title" content="{BRAND} — {$t('landing.tagline')}" />
	<meta property="og:description" content={$t('landing.subtitle')} />
	<meta property="og:type" content="website" />
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content="{BRAND} — {$t('landing.tagline')}" />
	<meta name="twitter:description" content={$t('landing.subtitle')} />
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<!-- Hero — editorial masthead: oversized display type, generous space. -->
		<section class="mx-auto w-full max-w-container px-4 pb-24 pt-20 md:px-12 md:pb-36 md:pt-32">
			<h1 class="max-w-5xl font-sans uppercase leading-[0.95] tracking-tighter text-display-xl">
				{$t('landing.tagline')}
			</h1>
			<p class="mt-10 max-w-2xl font-sans text-body-lg leading-relaxed text-text-muted">
				{$t('landing.subtitle')}
			</p>
			<div class="mt-12 flex flex-wrap items-center gap-4">
				<a href="/signup" class="flex items-center gap-3 btn-accent">
					{$t('landing.cta')}
					<ArrowRight size={16} />
				</a>
				<a href="/store/system-archive" class="btn-ghost">{$t('landing.seeDemo')}</a>
			</div>
		</section>

		<!-- Features — numbered editorial grid -->
		<section class="border-t border-border">
			<div class="mx-auto grid max-w-container grid-cols-1 gap-px bg-border md:grid-cols-2 lg:grid-cols-4">
				{#each features as key, i (key)}
					<div class="flex flex-col gap-4 bg-bg p-8 md:p-10">
						<span class="font-mono text-metadata-sm tracking-[0.1em] text-accent">
							{String(i + 1).padStart(2, '0')}
						</span>
						<h3 class="font-sans text-headline-md tracking-tight text-text">
							{$t('landing.' + key + 'Title')}
						</h3>
						<p class="font-sans text-body-md text-text-muted">
							{$t('landing.' + key + 'Body')}
						</p>
					</div>
				{/each}
			</div>
		</section>

		<!-- Core differentiator: no-oversell -->
		<section class="border-t border-border">
			<div class="mx-auto grid max-w-container grid-cols-1 gap-10 px-4 py-20 md:grid-cols-12 md:px-12">
				<div class="md:col-span-5">
					<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">
						<span class="text-text-muted">01 — </span>{$t('landing.coreKicker')}
					</p>
					<h2 class="mt-4 font-sans uppercase leading-none tracking-tighter text-headline-lg text-text">
						{$t('landing.coreTitle')}
					</h2>
				</div>
				<div class="md:col-span-7">
					<p class="font-sans text-body-lg leading-relaxed text-text-muted">
						{$t('landing.coreBody')}
					</p>
				</div>
			</div>
		</section>

		<!-- Live examples: the seeded demo stores, each its own theme -->
		<section class="border-t border-border">
			<div class="mx-auto max-w-container px-4 py-20 md:px-12">
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">
					<span class="text-text-muted">02 — </span>{$t('landing.examplesKicker')}
				</p>
				<h2 class="mt-4 max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg text-text">
					{$t('landing.examplesTitle')}
				</h2>
				<p class="mt-6 max-w-2xl font-sans text-body-lg text-text-muted">{$t('landing.examplesBody')}</p>
				<div class="mt-12 grid grid-cols-1 gap-px bg-border md:grid-cols-2 lg:grid-cols-4">
					{#each examples as ex (ex.slug)}
						<a
							href="/store/{ex.slug}"
							class="group flex flex-col gap-3 bg-bg p-8 transition-colors hover:bg-surface-variant"
						>
							<span class="font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">{ex.theme}</span>
							<h3 class="font-sans text-headline-md tracking-tight text-text">{ex.name}</h3>
							<p class="font-sans text-body-md text-text-muted">{ex.blurb}</p>
							<span class="mt-auto flex items-center gap-2 pt-4 font-mono text-metadata-sm uppercase tracking-[0.1em] text-accent">
								{$t('landing.seeDemo')}
								<ArrowRight size={14} class="transition-transform group-hover:translate-x-1" />
							</span>
						</a>
					{/each}
				</div>
			</div>
		</section>

		<!-- Pricing -->
		<section class="border-t border-border">
			<div class="mx-auto max-w-container px-4 py-20 md:px-12">
				<p class="mb-4 font-mono text-metadata-sm uppercase tracking-[0.2em] text-text-muted">03</p>
				<h2 class="mb-12 font-sans uppercase tracking-tighter text-headline-lg text-text">
					{$t('landing.pricingTitle')}
				</h2>
				<div class="grid grid-cols-1 gap-px bg-border md:grid-cols-3">
					{#each plans as plan (plan)}
						<div class="flex flex-col gap-6 bg-bg p-8">
							<div>
								<h3 class="font-sans text-headline-md tracking-tight text-text">
									{$t('landing.plan_' + plan + '_name')}
								</h3>
								<p class="mt-2 font-sans text-headline-lg tracking-tighter text-text">
									{$t('landing.plan_' + plan + '_price')}
								</p>
							</div>
							<p class="font-sans text-body-md text-text-muted">{$t('landing.plan_' + plan + '_desc')}</p>
							<a href="/signup" class="mt-auto {plan === 'pro' ? 'btn-accent' : 'btn-ghost'}">
								{$t('landing.cta')}
							</a>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<!-- Final CTA -->
		<section class="border-t border-border bg-inverse text-on-inverse">
			<div class="mx-auto flex max-w-container flex-col items-start gap-8 px-4 py-24 md:px-12">
				<h2 class="max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg">
					{$t('landing.finalTitle')}
				</h2>
				<a href="/signup" class="flex items-center gap-3 btn-accent">
					{$t('landing.cta')}
					<ArrowRight size={16} />
				</a>
			</div>
		</section>
	</main>

	<SiteFooter />
</div>
