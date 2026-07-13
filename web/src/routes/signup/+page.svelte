<script lang="ts">
	import { goto } from '$app/navigation';
	import { signup, createStore, selectStore, listStores } from '$lib/admin/api';
	import { BRAND, ROOT_DOMAIN } from '$lib/brand';
	import { t } from '$lib/i18n';
	import { THEMES } from '$lib/theme';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import type { ApiError } from '$lib/types';

	let email = $state('');
	let password = $state('');
	let storeName = $state('');
	let slug = $state('');
	let slugTouched = $state(false);
	// Default to a non-monolith theme so a store created without touching this picker still looks like its
	// own thing, not a clone of the TuTienda landing (which is monolith). Each theme brings its own accent.
	let theme = $state('boutique');
	let busy = $state(false);
	let error = $state('');
	let slugError = $state('');

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
		slugError = '';
		try {
			await signup(email, password);
		} catch (err) {
			error = (err as ApiError)?.error ?? $t('auth.failed');
			busy = false;
			return;
		}
		try {
			// Merchant account exists at this point. Create their first store (with the chosen theme so it
			// looks distinct from the start), then enter it → dashboard.
			const store = await createStore(effectiveSlug, storeName, theme);
			await selectStore(store);
			await goto('/app');
		} catch (err) {
			const apiErr = err as ApiError;
			if (apiErr?.error === 'slug already taken') {
				// The signup call right above just succeeded, so this merchant is real. If a store with
				// this exact slug already belongs to them — e.g. a prior submit's response got lost but the
				// store was actually created — pick it up instead of showing a confusing duplicate error.
				const mine = await listStores().catch(() => []);
				const existing = mine.find((s) => s.slug === effectiveSlug);
				if (existing) {
					await selectStore(existing);
					await goto('/app');
					return;
				}
				slugError = $t('auth.slugTaken');
			} else {
				error = apiErr?.error ?? $t('auth.failed');
			}
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
					{ROOT_DOMAIN}/store/{effectiveSlug || 'tu-tienda'}
				</span>
				{#if slugError}<span class="font-mono text-metadata-sm text-error">{slugError}</span>{/if}
			</label>

			<fieldset class="flex flex-col gap-2">
				<legend class="mb-1 {labelClass}">{$t('app.settings.theme')}</legend>
				<div class="grid grid-cols-2 gap-2">
					{#each THEMES as th (th.id)}
						<button
							type="button"
							onclick={() => (theme = th.id)}
							aria-pressed={theme === th.id}
							class="brutal-border flex flex-col gap-1 p-3 text-left transition-colors focus-ring
								{theme === th.id ? 'border-accent bg-surface' : 'bg-surface/40 hover:bg-surface'}"
						>
							<span class="font-sans text-body-md text-text">{th.label}</span>
							<span class="font-mono text-metadata-sm leading-snug text-text-muted">{th.description}</span>
						</button>
					{/each}
				</div>
			</fieldset>

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
