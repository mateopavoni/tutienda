<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { get } from 'svelte/store';
	import { login } from '$lib/admin/api';
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import type { ApiError } from '$lib/types';

	let email = $state('');
	let password = $state('');
	let busy = $state(false);
	let error = $state($page.url.searchParams.get('expired') ? get(t)('auth.sessionExpired') : '');

	const inputClass = 'brutal-border min-w-0 bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			await login(email, password);
			await goto('/app');
		} catch (err) {
			const e = err as ApiError;
			error = e?.error ?? $t('auth.failed');
			// A failed login never wipes what was typed — the usual case is a wrong password, and making
			// someone retype their email for that is pure friction. The one exception is an address with
			// no account at all (clear_email from the API), where the email is the thing to fix. The
			// message the user reads is the same either way; only the field differs.
			if (e?.clear_email) email = '';
		} finally {
			busy = false;
		}
	}

	async function tryDemo() {
		busy = true;
		error = '';
		try {
			await login('demo@system-archive.store', 'demo-archive-2026');
			await goto('/app');
		} catch (err) {
			error = (err as ApiError)?.error ?? $t('auth.failed');
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{BRAND} — {$t('auth.login')}</title></svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader variant="minimal" />
	<main class="mx-auto flex w-full max-w-md flex-1 flex-col justify-center px-4 py-16">
		<h1 class="mb-2 font-sans text-headline-lg tracking-tighter">{$t('auth.loginTitle')}</h1>
		<p class="mb-8 {labelClass}">{$t('auth.loginSubtitle')}</p>

		<button
			type="button"
			onclick={tryDemo}
			disabled={busy}
			class="btn-primary w-full py-4 text-body-lg disabled:opacity-50"
		>
			{busy ? '…' : $t('auth.tryDemo')}
		</button>
		<p class="mt-2 mb-8 {labelClass}">{$t('auth.tryDemoHint')}</p>

		<form onsubmit={submit} class="flex flex-col gap-4">
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.email')}</span>
				<input type="email" bind:value={email} required class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.password')}</span>
				<PasswordInput bind:value={password} required minlength={8} class={inputClass} />
			</label>

			{#if error}<p class="font-mono text-metadata-sm text-error">{error}</p>{/if}

			<button type="submit" disabled={busy} class="btn-ghost mt-2 disabled:opacity-50">
				{busy ? '…' : $t('auth.login')}
			</button>
		</form>

		<a href="/signup" class="mt-6 {labelClass} transition-colors hover:text-text">
			{$t('auth.toSignup')}
		</a>
	</main>
</div>
