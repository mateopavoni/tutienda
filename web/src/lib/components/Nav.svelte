<script lang="ts">
	import { t } from '$lib/i18n';
	import { cartCount } from '$lib/stores/cart';
	import { cartOpen } from '$lib/stores/ui';
	import { Settings } from 'lucide-svelte';
	import type { StoreInfo } from '$lib/tenant';

	let { store, slug }: { store?: StoreInfo | null; slug: string } = $props();
	// Per-store wordmark; the type/scale stay fixed by the design system.
	let wordmark = $derived((store?.displayName ?? slug).toUpperCase());
	let home = $derived('/store/' + slug);
</script>

<header class="fixed top-0 z-50 w-full border-b border-border bg-bg">
	<div class="mx-auto flex h-16 max-w-container items-center justify-between px-4 md:px-12">
		<a href={home} class="flex items-center gap-2 font-sans text-headline-md tracking-tighter text-text">
			{#if store?.settings?.logoUrl}
				<img src={store.settings.logoUrl} alt={wordmark} class="h-7 w-auto" />
			{:else}
				{wordmark}
			{/if}
		</a>

		<div class="flex items-center gap-4 md:gap-6">
			<a
				href={home}
				class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
			>
				{$t('nav.catalog')}
			</a>
			<a
				href="/configuracion"
				class="p-2 text-text-muted transition-colors hover:text-text"
				aria-label={$t('settings.title')}
				title={$t('settings.title')}
			>
				<Settings size={18} strokeWidth={1.5} />
			</a>
			<button
				onclick={() => cartOpen.set(true)}
				class="font-sans text-label-caps uppercase tracking-[0.1em] text-text transition-colors hover:text-accent"
			>
				{$t('nav.bag')} ({$cartCount})
			</button>
		</div>
	</div>
</header>
