<script lang="ts">
	// About the platform (marketing). Distinct from a merchant's own storefront "about" content, which
	// doesn't exist anymore — see $lib/pages (only Terms remains per-store).
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { ArrowRight } from 'lucide-svelte';
	import { fly } from 'svelte/transition';

	const principles = ['principle1', 'principle2', 'principle3'];
</script>

<svelte:head>
	<title>{BRAND} — {$t('marketing.about.title')}</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<section class="mx-auto w-full max-w-container px-4 pb-20 pt-20 md:px-12 md:pb-28 md:pt-32">
			<div in:fly={{ y: 16, duration: 400 }}>
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.about.kicker')}</p>
				<h1 class="mt-4 max-w-4xl font-sans uppercase leading-[0.95] tracking-tighter text-display-xl">
					{$t('marketing.about.title')}
				</h1>
				<p class="mt-8 max-w-2xl font-sans text-body-lg leading-relaxed text-text-muted">
					{$t('marketing.about.intro', { brand: BRAND })}
				</p>
			</div>
		</section>

		<section class="border-t border-border">
			<div class="mx-auto grid max-w-container grid-cols-1 gap-10 px-4 py-20 md:grid-cols-12 md:px-12">
				<div class="md:col-span-5" in:fly={{ y: 16, duration: 400, delay: 50 }}>
					<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.about.missionKicker')}</p>
					<h2 class="mt-4 font-sans uppercase leading-none tracking-tighter text-headline-lg text-text">
						{$t('marketing.about.missionTitle')}
					</h2>
				</div>
				<div class="md:col-span-7" in:fly={{ y: 16, duration: 400, delay: 100 }}>
					<p class="font-sans text-body-lg leading-relaxed text-text-muted">
						{$t('marketing.about.missionBody', { brand: BRAND })}
					</p>
				</div>
			</div>
		</section>

		<section class="border-t border-border">
			<div class="mx-auto max-w-container px-4 py-20 md:px-12">
				<div in:fly={{ y: 16, duration: 400 }}>
					<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.about.principlesKicker')}</p>
					<h2 class="mt-4 max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg text-text">
						{$t('marketing.about.principlesTitle')}
					</h2>
				</div>
				<div class="mt-12 grid grid-cols-1 gap-px bg-border md:grid-cols-3">
					{#each principles as key, i (key)}
						<div
							class="flex flex-col gap-4 bg-bg p-8"
							in:fly={{ y: 16, duration: 400, delay: 100 + i * 80 }}
						>
							<span class="font-mono text-metadata-sm tracking-[0.1em] text-accent">{String(i + 1).padStart(2, '0')}</span>
							<h3 class="font-sans text-headline-md tracking-tight text-text">
								{$t('marketing.about.' + key + 'Title')}
							</h3>
							<p class="font-sans text-body-md text-text-muted">
								{$t('marketing.about.' + key + 'Body')}
							</p>
						</div>
					{/each}
				</div>
			</div>
		</section>

		<section class="border-t border-border bg-inverse text-on-inverse">
			<div class="mx-auto flex max-w-container flex-col items-start gap-8 px-4 py-24 md:px-12">
				<h2 class="max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg">
					{$t('marketing.about.ctaTitle')}
				</h2>
				<div class="flex flex-wrap items-center gap-6">
					<a href="/store/system-archive" class="flex items-center gap-3 btn-accent">
						{$t('marketing.about.cta')}
						<ArrowRight size={16} />
					</a>
					<a href="/signup" class="font-sans text-label-caps uppercase tracking-[0.1em] underline-offset-4 hover:underline">
						{$t('marketing.about.ctaSecondary')}
					</a>
				</div>
			</div>
		</section>
	</main>

	<SiteFooter />
</div>
