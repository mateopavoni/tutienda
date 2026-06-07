<script lang="ts">
	// Marketing/platform header (landing, auth, settings). Distinct from the storefront Nav: it sells the
	// SaaS, not a store. Brand name is centralized in $lib/brand.
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import { Settings } from 'lucide-svelte';

	let { variant = 'full' }: { variant?: 'full' | 'minimal' } = $props();
</script>

<header class="w-full border-b border-border bg-bg">
	<div class="mx-auto flex h-16 max-w-container items-center justify-between px-4 md:px-12">
		<a href="/" class="font-sans text-headline-md tracking-tighter text-text">{BRAND}</a>

		<div class="flex items-center gap-4 md:gap-6">
			{#if variant === 'full'}
				<a
					href="/login"
					class="font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
				>
					{$t('auth.login')}
				</a>
				<a href="/signup" class="btn-primary">{$t('landing.cta')}</a>
			{/if}
			<a
				href="/configuracion"
				class="p-2 text-text-muted transition-colors hover:text-text"
				aria-label={$t('settings.title')}
				title={$t('settings.title')}
			>
				<Settings size={18} strokeWidth={1.5} />
			</a>
		</div>
	</div>
</header>
