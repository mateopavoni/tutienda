<script lang="ts">
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { customerLogin } from '$lib/api/customer';
	import { setCustomerSession } from '$lib/stores/customer';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import type { ApiError } from '$lib/types';

	let { data } = $props();
	const slug = $derived(data.storeSlug);
	const storeName = $derived(data.store?.displayName ?? data.storeSlug);
	const accountHome = $derived('/store/' + slug + '/cuenta');

	let email = $state('');
	let password = $state('');
	let busy = $state(false);
	let error = $state('');

	const inputClass = 'brutal-border min-w-0 bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			const s = await customerLogin(email, password);
			setCustomerSession(slug, s.token, s.customer);
			await goto(accountHome);
		} catch (err) {
			const e = err as ApiError;
			error = e?.error ?? $t('account.failed');
			// Same rule as the merchant login: keep the email so a wrong password only costs a retype of
			// the password, and blank it only when the API says no account exists for it (clear_email).
			// The error text above is identical in both cases.
			if (e?.clear_email) email = '';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{storeName} — {$t('account.loginTitle')}</title></svelte:head>

<section class="mx-auto flex w-full max-w-md flex-col px-4 py-16 md:py-24">
	<h1 class="mb-2 font-sans text-headline-lg tracking-tighter">{$t('account.loginTitle')}</h1>
	<p class="mb-8 {labelClass}">{$t('account.subtitle', { store: storeName })}</p>

	<form onsubmit={submit} class="flex flex-col gap-4">
		<label class="flex flex-col gap-1">
			<span class={labelClass}>{$t('account.email')}</span>
			<input type="email" bind:value={email} required class={inputClass} />
		</label>
		<label class="flex flex-col gap-1">
			<span class={labelClass}>{$t('account.password')}</span>
			<PasswordInput bind:value={password} required class={inputClass} />
		</label>

		{#if error}<p class="font-mono text-metadata-sm text-error">{error}</p>{/if}

		<button type="submit" disabled={busy} class="btn-accent mt-2 disabled:opacity-50">
			{busy ? '…' : $t('account.signInCta')}
		</button>
	</form>

	<a href="{accountHome}/registro" class="mt-6 {labelClass} transition-colors hover:text-text">
		{$t('account.toRegister')}
	</a>
	<p class="mt-8 font-sans text-body-md text-text-muted">{$t('account.guestNote')}</p>
</section>
