<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/admin/api';
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import type { ApiError } from '$lib/types';

	let email = $state('');
	let password = $state('');
	let busy = $state(false);
	let error = $state('');

	const inputClass = 'brutal-border bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			await login(email, password);
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

		<form onsubmit={submit} class="flex flex-col gap-4">
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.email')}</span>
				<input type="email" bind:value={email} required class={inputClass} />
			</label>
			<label class="flex flex-col gap-1">
				<span class={labelClass}>{$t('auth.password')}</span>
				<input type="password" bind:value={password} required minlength="8" class={inputClass} />
			</label>

			{#if error}<p class="font-mono text-metadata-sm text-error">{error}</p>{/if}

			<button type="submit" disabled={busy} class="btn-primary mt-2 disabled:opacity-50">
				{busy ? '…' : $t('auth.login')}
			</button>
		</form>

		<a href="/signup" class="mt-6 {labelClass} transition-colors hover:text-text">
			{$t('auth.toSignup')}
		</a>
	</main>
</div>
