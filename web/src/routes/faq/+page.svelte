<script lang="ts">
	// Platform FAQ (marketing). Same accordion pattern as the storefront's FAQ section
	// (StorefrontSections.svelte) — native <details>/<summary>, no JS needed for the toggle.
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { ChevronDown } from 'lucide-svelte';
	import { fly } from 'svelte/transition';

	const questions = ['q1', 'q2', 'q3', 'q4', 'q5', 'q6', 'q7', 'q8'];
</script>

<svelte:head>
	<title>{BRAND} — {$t('marketing.faq.title')}</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<section class="mx-auto w-full max-w-container px-4 pb-12 pt-20 md:px-12 md:pt-32">
			<div in:fly={{ y: 16, duration: 400 }}>
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.faq.kicker')}</p>
				<h1 class="mt-4 max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg md:text-display-xl">
					{$t('marketing.faq.title')}
				</h1>
				<p class="mt-6 max-w-2xl font-sans text-body-lg text-text-muted">{$t('marketing.faq.subtitle')}</p>
			</div>
		</section>

		<section class="border-t border-border">
			<div class="mx-auto max-w-3xl px-4 py-12 md:px-12">
				<div class="flex flex-col divide-y divide-border border-b border-border">
					{#each questions as key, i (key)}
						<div in:fly={{ y: 12, duration: 300, delay: i * 40 }}>
							<details class="group py-6">
								<summary class="flex cursor-pointer list-none items-center justify-between gap-4 font-sans text-body-lg text-text">
									{$t('marketing.faq.' + key + 'Title', { brand: BRAND })}
									<ChevronDown size={18} class="shrink-0 text-text-muted transition-transform group-open:rotate-180" />
								</summary>
								<p class="mt-4 font-sans text-body-md leading-relaxed text-text-muted">
									{$t('marketing.faq.' + key + 'Body', { brand: BRAND })}
								</p>
							</details>
						</div>
					{/each}
				</div>
			</div>
		</section>
	</main>

	<SiteFooter />
</div>
