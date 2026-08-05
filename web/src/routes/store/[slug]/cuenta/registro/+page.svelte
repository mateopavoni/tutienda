<script lang="ts">
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { customerSignup } from '$lib/api/customer';
	import { setCustomerSession } from '$lib/stores/customer';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import type { ApiError } from '$lib/types';

	let { data } = $props();
	const slug = $derived(data.storeSlug);
	const storeName = $derived(data.store?.displayName ?? data.storeSlug);
	const accountHome = $derived('/store/' + slug + '/cuenta');

	let email = $state('');
	let password = $state('');
	let name = $state('');
	let busy = $state(false);
	let error = $state('');

	const inputClass = 'brutal-border min-w-0 bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		busy = true;
		error = '';
		try {
			const s = await customerSignup(email, password, name);
			setCustomerSession(slug, s.token, s.customer);
			await goto(accountHome);
		} catch (err) {
			error = (err as ApiError)?.error ?? $t('account.failed');
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>{storeName} — {$t('account.registerTitle')}</title></svelte:head>

<section class="mx-auto flex w-full max-w-md flex-col px-4 py-16 md:py-24">
	<h1 class="mb-2 font-sans text-headline-lg tracking-tighter">{$t('account.registerTitle')}</h1>
	<p class="mb-8 {labelClass}">{$t('account.registerSubtitle', { store: storeName })}</p>

	<form onsubmit={submit} class="flex flex-col gap-4">
		<label class="flex flex-col gap-1">
			<span class={labelClass}>{$t('account.nameOptional')}</span>
			<input bind:value={name} class={inputClass} />
		</label>
		<label class="flex flex-col gap-1">
			<span class={labelClass}>{$t('account.email')}</span>
			<input type="email" bind:value={email} required class={inputClass} />
		</label>
		<label class="flex flex-col gap-1">
			<span class={labelClass}>{$t('account.password')}</span>
			<PasswordInput
				bind:value={password}
				required
				minlength={8}
				autocomplete="new-password"
				class={inputClass}
			/>
		</label>

		{#if error}<p class="font-mono text-metadata-sm text-error">{error}</p>{/if}

		<button type="submit" disabled={busy} class="btn-accent mt-2 disabled:opacity-50">
			{busy ? '…' : $t('account.registerCta')}
		</button>
	</form>

	<a href="{accountHome}/login" class="mt-6 {labelClass} transition-colors hover:text-text">
		{$t('account.toLogin')}
	</a>
</section>
