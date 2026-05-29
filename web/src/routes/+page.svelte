<script lang="ts">
	// Marketing landing for the SaaS. This is NOT a storefront — it sells the platform. The actual stores
	// live under /store/[slug] (subdomains in production). Brand name comes from $lib/brand.
	import { BRAND, ROOT_DOMAIN } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { ArrowRight, ShieldCheck, Zap, Palette, Store } from 'lucide-svelte';

	const features = [
		{ icon: Store, key: 'feat1' },
		{ icon: Palette, key: 'feat2' },
		{ icon: ShieldCheck, key: 'feat3' },
		{ icon: Zap, key: 'feat4' }
	];

	const plans = ['free', 'pro', 'scale'] as const;
</script>

<svelte:head>
	<title>{BRAND} — {$t('landing.tagline')}</title>
	<meta name="description" content={$t('landing.subtitle')} />
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<!-- Hero -->
		<section class="mx-auto w-full max-w-container px-4 pb-20 pt-16 md:px-12 md:pb-28 md:pt-24">
			<p class="mb-6 font-mono text-metadata-sm uppercase tracking-[0.2em] text-text-muted">
				{$t('landing.kicker')}
			</p>
			<h1 class="max-w-5xl font-sans uppercase leading-none tracking-tighter text-display-xl">
				{$t('landing.tagline')}
			</h1>
			<p class="mt-8 max-w-2xl font-sans text-body-lg text-text-muted">
				{$t('landing.subtitle')}
			</p>
			<div class="mt-10 flex flex-wrap items-center gap-4">
				<a href="/signup" class="flex items-center gap-3 btn-accent">
					{$t('landing.cta')}
					<ArrowRight size={16} />
				</a>
				<a href="/store/system-archive" class="btn-ghost">{$t('landing.seeDemo')}</a>
			</div>
			<p class="mt-6 font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">
				{$t('landing.subdomainNote', { domain: ROOT_DOMAIN })}
			</p>
		</section>

		<!-- Features -->
		<section class="border-t border-border">
			<div class="mx-auto grid max-w-container grid-cols-1 gap-px bg-border md:grid-cols-2 lg:grid-cols-4">
				{#each features as feature (feature.key)}
					<div class="flex flex-col gap-4 bg-bg p-8 md:p-10">
						<feature.icon size={24} strokeWidth={1.5} class="text-accent" />
						<h3 class="font-sans text-headline-md tracking-tight text-text">
							{$t('landing.' + feature.key + 'Title')}
						</h3>
						<p class="font-sans text-body-md text-text-muted">
							{$t('landing.' + feature.key + 'Body')}
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
						{$t('landing.coreKicker')}
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

		<!-- Pricing -->
		<section class="border-t border-border">
			<div class="mx-auto max-w-container px-4 py-20 md:px-12">
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
