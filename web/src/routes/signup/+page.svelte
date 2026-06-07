<script lang="ts">
	import { goto } from '$app/navigation';
	import { signup, createStore, selectStore } from '$lib/admin/api';
	import { BRAND, ROOT_DOMAIN } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import type { ApiError } from '$lib/types';

	let email = $state('');
	let password = $state('');
	let storeName = $state('');
	let slug = $state('');
	let slugTouched = $state(false);
	let busy = $state(false);
	let error = $state('');

	// Onboarding doubles as store creation: a slug is derived from the store name until the user edits it.
	const slugify = (s: string) =>
		s.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 40);
	const effectiveSlug = $derived(slugTouched ? slug : slugify(storeName));

	const inputClass = 'brutal-border bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			// 1) create the merchant account, 2) create their first store, 3) enter it → dashboard.
			await signup(email, password);
			const store = await createStore(effectiveSlug, storeName);
			await selectStore(store);
			await goto('/app');
		} catch (err) {
			error = (err as ApiError)?.error ?? $t('auth.failed');
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{BRAND} — {$t('auth.signupTitle')}</title></svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader variant="minimal" />
	<main class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center px-4 py-16">
		<h1 class="mb-2 font-sans text-headline-lg tracking-tighter">{$t('auth.signupTitle')}</h1>
		<p class="mb-8 {labelClass}">{$t('auth.signupSubtitle')}</p>

		<form onsubmit={submit} class="flex flex-col gap-4">
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.email')}</span>
				<input type="email" bind:value={email} required class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.password')}</span>
				<input type="password" bind:value={password} required minlength="8" class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.storeName')}</span>
				<input bind:value={storeName} required placeholder="ACME" class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.storeSlug')}</span>
				<input
					value={effectiveSlug}
					oninput={(e) => {
						slugTouched = true;
						slug = e.currentTarget.value;
					}}
					required
					placeholder="acme"
					class={inputClass}
				/>
				<span class="font-mono text-metadata-sm text-text-muted">
					{effectiveSlug || 'tu-tienda'}.{ROOT_DOMAIN}
				</span>
			</label>

			{#if error}<p class="font-mono text-metadata-sm text-error">{error}</p>{/if}

			<button type="submit" disabled={busy} class="btn-accent mt-2 disabled:opacity-50">
				{busy ? '…' : $t('landing.cta')}
			</button>
		</form>

		<a href="/login" class="mt-6 {labelClass} transition-colors hover:text-text">
			{$t('auth.toLogin')}
		</a>
	</main>
</div>
