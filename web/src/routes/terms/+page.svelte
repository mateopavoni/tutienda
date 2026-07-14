<script lang="ts">
	// Platform Terms of Service (marketing) — using TuTienda as a merchant, distinct from the per-store
	// Terms each merchant writes for their own customers (see $lib/pages, /store/[slug]/terms).
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { fly } from 'svelte/transition';

	const sections = ['s1', 's2', 's3', 's4', 's5', 's6', 's7', 's8', 's9'];
</script>

<svelte:head>
	<title>{BRAND} — {$t('marketing.terms.title')}</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<section class="mx-auto w-full max-w-3xl px-4 pb-12 pt-20 md:px-12 md:pt-32">
			<div in:fly={{ y: 16, duration: 400 }}>
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.terms.kicker')}</p>
				<h1 class="mt-4 font-sans uppercase leading-none tracking-tighter text-headline-lg">
					{$t('marketing.terms.title')}
				</h1>
				<p class="mt-4 font-mono text-metadata-sm text-text-muted">{$t('marketing.terms.updated')}</p>
				<p class="mt-8 font-sans text-body-lg leading-relaxed text-text-muted">
					{$t('marketing.terms.intro', { brand: BRAND })}
				</p>
			</div>
		</section>

		<section class="border-t border-border">
			<div class="mx-auto max-w-3xl px-4 py-12 md:px-12">
				<div class="flex flex-col divide-y divide-border">
					{#each sections as key, i (key)}
						<div class="py-8" in:fly={{ y: 12, duration: 300, delay: i * 30 }}>
							<h2 class="font-sans text-headline-md tracking-tight text-text">
								{$t('marketing.terms.' + key + 'Heading')}
							</h2>
							<p class="mt-3 font-sans text-body-md leading-relaxed text-text-muted">
								{$t('marketing.terms.' + key + 'Body', { brand: BRAND })}
							</p>
						</div>
					{/each}
				</div>
			</div>
		</section>
	</main>

	<SiteFooter />
</div>
