<script lang="ts">
	// Marketing/platform header (landing, auth, settings). Distinct from the storefront Nav: it sells the
	// SaaS, not a store. Brand name is centralized in $lib/brand.
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import { Settings, Menu, X } from 'lucide-svelte';
	import { session } from '$lib/admin/session';

	let { variant = 'full' }: { variant?: 'full' | 'minimal' } = $props();
	// Below md, the About/FAQ/Contact/login links are hidden (see md:inline below) with no other way to
	// reach them — this panel is that mobile replacement, same pattern as the storefront Nav.
	let mobileOpen = $state(false);
	// localStorage only exists client-side; starts '' to match SSR, $effect fills it in after mount so a
	// signed-in merchant sees "Tus tiendas" instead of Ingresar/Crear tu tienda on every marketing page.
	let merchantToken = $state('');
	$effect(() => {
		merchantToken = session.merchantToken();
	});
</script>

<header class="w-full border-b border-border bg-bg">
	<div class="mx-auto flex h-16 max-w-container items-center justify-between gap-3 px-4 md:px-12">
		<a href="/" class="truncate font-sans text-body-lg tracking-tighter text-text md:text-headline-md">{BRAND}</a>

		<div class="flex items-center gap-3 md:gap-6">
			{#if variant === 'full'}
				<a
					href="/about"
					class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
				>
					{$t('nav.about')}
				</a>
				<a
					href="/faq"
					class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
				>
					{$t('nav.faq')}
				</a>
				<a
					href="/contact"
					class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text md:inline"
				>
					{$t('nav.contact')}
				</a>
				{#if merchantToken}
					<a href="/app" class="btn-primary px-4 py-2.5 md:px-6 md:py-4">{$t('nav.yourStores')}</a>
				{:else}
					<a
						href="/login"
						class="hidden font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text sm:inline"
					>
						{$t('auth.login')}
					</a>
					<a href="/signup" class="btn-primary px-4 py-2.5 md:px-6 md:py-4">{$t('landing.cta')}</a>
				{/if}
			{/if}
			<a
				href="/configuracion"
				class="p-2 text-text-muted transition-colors hover:text-text"
				aria-label={$t('settings.title')}
				title={$t('settings.title')}
			>
				<Settings size={18} strokeWidth={1.5} />
			</a>
			{#if variant === 'full'}
				<button
					onclick={() => (mobileOpen = !mobileOpen)}
					class="p-2 text-text-muted transition-colors hover:text-text md:hidden"
					aria-label={$t('nav.menu')}
					aria-expanded={mobileOpen}
				>
					{#if mobileOpen}<X size={18} strokeWidth={1.5} />{:else}<Menu size={18} strokeWidth={1.5} />{/if}
				</button>
			{/if}
		</div>
	</div>
	{#if variant === 'full' && mobileOpen}
		<div class="flex flex-col gap-px border-t border-border bg-border md:hidden">
			<a
				href="/about"
				onclick={() => (mobileOpen = false)}
				class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
			>
				{$t('nav.about')}
			</a>
			<a
				href="/faq"
				onclick={() => (mobileOpen = false)}
				class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
			>
				{$t('nav.faq')}
			</a>
			<a
				href="/contact"
				onclick={() => (mobileOpen = false)}
				class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
			>
				{$t('nav.contact')}
			</a>
			{#if merchantToken}
				<a
					href="/app"
					onclick={() => (mobileOpen = false)}
					class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
				>
					{$t('nav.yourStores')}
				</a>
			{:else}
				<a
					href="/login"
					onclick={() => (mobileOpen = false)}
					class="bg-bg px-4 py-3 font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted transition-colors hover:text-text"
				>
					{$t('auth.login')}
				</a>
			{/if}
		</div>
	{/if}
</header>
