<script lang="ts">
	// Platform contact page (marketing). No form/backend — an honest mailto link instead of a contact
	// form that would silently go nowhere (there's no endpoint to receive it).
	import { BRAND, ROOT_DOMAIN } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { Mail, Clock, ArrowRight } from 'lucide-svelte';
	import { fly } from 'svelte/transition';

	const email = 'hola@' + ROOT_DOMAIN;
</script>

<svelte:head>
	<title>{BRAND} — {$t('marketing.contact.title')}</title>
</svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader />

	<main class="flex-1">
		<section class="mx-auto w-full max-w-container px-4 pb-20 pt-20 md:px-12 md:pb-28 md:pt-32">
			<div in:fly={{ y: 16, duration: 400 }}>
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">{$t('marketing.contact.kicker')}</p>
				<h1 class="mt-4 max-w-3xl font-sans uppercase leading-none tracking-tighter text-headline-lg md:text-display-xl">
					{$t('marketing.contact.title')}
				</h1>
				<p class="mt-6 max-w-2xl font-sans text-body-lg text-text-muted">{$t('marketing.contact.subtitle')}</p>
			</div>
		</section>

		<section class="border-t border-border">
			<div class="mx-auto grid max-w-container grid-cols-1 gap-px bg-border md:grid-cols-2">
				<div class="flex flex-col gap-3 bg-bg p-8 md:p-10" in:fly={{ y: 16, duration: 400, delay: 60 }}>
					<Mail size={20} strokeWidth={1.5} class="text-accent" />
					<span class="font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">
						{$t('marketing.contact.emailLabel')}
					</span>
					<a href="mailto:{email}" class="font-sans text-headline-md tracking-tight text-text hover:text-accent">
						{email}
					</a>
					<p class="font-sans text-body-md text-text-muted">{$t('marketing.contact.emailNote')}</p>
				</div>
				<div class="flex flex-col gap-3 bg-bg p-8 md:p-10" in:fly={{ y: 16, duration: 400, delay: 120 }}>
					<Clock size={20} strokeWidth={1.5} class="text-accent" />
					<span class="font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">
						{$t('marketing.contact.hoursLabel')}
					</span>
					<span class="font-sans text-headline-md tracking-tight text-text">{$t('marketing.contact.hoursBody')}</span>
				</div>
			</div>
		</section>

		<section class="border-t border-border bg-inverse text-on-inverse">
			<div class="mx-auto flex max-w-container flex-col items-start gap-6 px-4 py-20 md:px-12">
				<span class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-text-muted">
					{$t('marketing.contact.demoLabel')}
				</span>
				<p class="max-w-xl font-sans text-body-lg leading-relaxed">{$t('marketing.contact.demoBody')}</p>
				<a href="/store/system-archive" class="flex items-center gap-3 btn-accent">
					{$t('marketing.contact.demoCta')}
					<ArrowRight size={16} />
				</a>
			</div>
		</section>
	</main>

	<SiteFooter />
</div>
