<script lang="ts">
	import { t } from '$lib/i18n';
	import { cartCount } from '$lib/stores/cart';
	import { cartOpen } from '$lib/stores/ui';
	import { customer } from '$lib/stores/customer';
	import { Settings, User, Menu, X } from 'lucide-svelte';
	import type { StoreInfo } from '$lib/tenant';
	import { pageTitle, pageSlug, type Page } from '$lib/pages';

	let { store, slug, pages = [] }: { store?: StoreInfo | null; slug: string; pages?: Page[] } = $props();
	// Per-store wordmark; the type/scale stay fixed by the design system.
	let wordmark = $derived((store?.displayName ?? slug).toUpperCase());
	let home = $derived('/store/' + slug);
	let catalog = $derived(home + '/catalogo');
	// Account link points to the profile when signed in, otherwise to the login page.
	let account = $derived(home + '/cuenta' + ($customer ? '' : '/login'));
	// Below md, the Home/Catalog/pages links are hidden (see md:inline below) with no other way to reach
	// them — this panel is that mobile replacement.
	let mobileOpen = $state(false);
</script>

<header class="fixed top-0 z-50 w-full border-b border-border bg-bg">
	<div class="mx-auto flex h-16 max-w-container items-center justify-between px-4 md:px-12">
		<a href={home} class="flex items-center gap-2 truncate font-sans text-body-lg tracking-tighter text-text md:text-headline-md">
			{#if store?.settings?.logoUrl}
				<img src={store.settings.logoUrl} alt={wordmark} class="h-6 w-auto md:h-7" />
			{:else}
				{wordmark}
			{/if}
		</a>

		<div class="flex items-center gap-4 md:gap-6">
			<a
				href={home}
				class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
			>
				{$t('nav.home')}
			</a>
			<a
				href={catalog}
				class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
			>
				{$t('nav.catalog')}
			</a>
			{#each pages as p (p.type)}
				<a
					href="{home}/{pageSlug(p)}"
					class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
				>
					{pageTitle(p)}
				</a>
			{/each}
			<a
				href={account}
				class="p-2 transition-colors hover:text-text {$customer ? 'text-text' : 'text-text-muted'}"
				aria-label={$t('account.title')}
				title={$t('account.title')}
			>
				<User size={18} strokeWidth={1.5} />
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
			<button
				onclick={() => (mobileOpen = !mobileOpen)}
				class="p-2 text-text-muted transition-colors hover:text-text md:hidden"
				aria-label={$t('nav.menu')}
				aria-expanded={mobileOpen}
			>
				{#if mobileOpen}<X size={18} strokeWidth={1.5} />{:else}<Menu size={18} strokeWidth={1.5} />{/if}
			</button>
		</div>
	</div>
	{#if mobileOpen}
		<div class="flex flex-col gap-px border-t border-border bg-border md:hidden">
			<a
				href={home}
				onclick={() => (mobileOpen = false)}
				class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
			>
				{$t('nav.home')}
			</a>
			<a
				href={catalog}
				onclick={() => (mobileOpen = false)}
				class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
			>
				{$t('nav.catalog')}
			</a>
			{#each pages as p (p.type)}
				<a
					href="{home}/{pageSlug(p)}"
					onclick={() => (mobileOpen = false)}
					class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
				>
					{pageTitle(p)}
				</a>
			{/each}
		</div>
	{/if}
</header>
