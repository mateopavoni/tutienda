<script lang="ts">
	// Preferences live here, not in any navbar (deliberate UX decision): language + light/dark/system.
	import { BRAND } from '$lib/brand';
	import { t, locale, locales, type Locale } from '$lib/i18n';
	import { theme, type Theme } from '$lib/stores/theme';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import { Sun, Moon, Monitor } from 'lucide-svelte';

	const themes: { value: Theme; icon: typeof Sun }[] = [
		{ value: 'system', icon: Monitor },
		{ value: 'light', icon: Sun },
		{ value: 'dark', icon: Moon }
	];

	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted';
</script>

<svelte:head><title>{BRAND} — {$t('settings.title')}</title></svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader variant="minimal" />
	<main class="mx-auto w-full max-w-container flex-1 px-4 py-16 md:px-12">
		<h1 class="mb-12 font-sans text-headline-lg tracking-tighter">{$t('settings.title')}</h1>

		<div class="grid grid-cols-1 gap-12 md:max-w-2xl">
			<!-- Language -->
			<section class="flex flex-col gap-4">
				<h2 class={labelClass}>{$t('settings.language')}</h2>
				<div class="flex gap-px brutal-border bg-border">
					{#each locales as option (option.code)}
						<button
							onclick={() => locale.set(option.code as Locale)}
							class="flex-1 px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] transition-colors
								{$locale === option.code ? 'bg-primary text-on-primary' : 'bg-bg text-text-muted hover:text-text'}"
							aria-pressed={$locale === option.code}
						>
							{option.label}
						</button>
					{/each}
				</div>
			</section>

			<!-- Theme -->
			<section class="flex flex-col gap-4">
				<h2 class={labelClass}>{$t('settings.theme')}</h2>
				<div class="flex gap-px brutal-border bg-border">
					{#each themes as option (option.value)}
						<button
							onclick={() => theme.set(option.value)}
							class="flex flex-1 items-center justify-center gap-2 px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] transition-colors
								{$theme === option.value ? 'bg-primary text-on-primary' : 'bg-bg text-text-muted hover:text-text'}"
							aria-pressed={$theme === option.value}
						>
							<option.icon size={16} strokeWidth={1.5} />
							{$t('theme.' + option.value)}
						</button>
					{/each}
				</div>
			</section>
		</div>
	</main>
</div>
