<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { t, locale } from '$lib/i18n';
	import { money } from '$lib/format';
	import { myOrders } from '$lib/api/customer';
	import { customer, customerToken, loadCustomer, clearCustomer } from '$lib/stores/customer';
	import { FULFILLMENT_STAGES, fulfillmentIndex } from '$lib/order';
	import type { ApiError, Order } from '$lib/types';

	let { data } = $props();
	const slug = $derived(data.storeSlug);
	const storeName = $derived(data.store?.displayName ?? data.storeSlug);

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let failed = $state(false);

	onMount(async () => {
		// Hydrate the session for this store (idempotent) before reading the token, independent of the
		// layout's own hydration order. No token ⇒ bounce to login.
		loadCustomer(slug);
		const token = customerToken();
		if (!token) {
			await goto('/store/' + slug + '/cuenta/login');
			return;
		}
		try {
			orders = await myOrders(token);
		} catch (err) {
			// A 401 means the session is stale/mismatched (not a real failure to load) — clear it and send
			// the customer back to login instead of showing a generic "couldn't load" error.
			if ((err as ApiError).status === 401) {
				clearCustomer(slug);
				await goto('/store/' + slug + '/cuenta/login');
				return;
			}
			failed = true;
		} finally {
			loading = false;
		}
	});

	function signOut() {
		clearCustomer(slug);
		goto('/store/' + slug);
	}

	// Tick so the simulated fulfillment stage of recent orders advances on screen.
	let now = $state(Date.now());
	$effect(() => {
		const id = setInterval(() => (now = Date.now()), 2000);
		return () => clearInterval(id);
	});

	// A confirmed order shows its live simulated stage; otherwise the raw status (e.g. FAILED).
	function statusLabel(order: Order): string {
		const i = fulfillmentIndex(order, now);
		return i >= 0 ? $t('cart.stage_' + FULFILLMENT_STAGES[i]) : order.status;
	}

	const fmtDate = (iso: string) => new Date(iso).toLocaleDateString($locale);
</script>

<svelte:head><title>{storeName} — {$t('account.title')}</title></svelte:head>

<section class="mx-auto w-full max-w-3xl px-4 py-16 md:py-24">
	<header class="mb-12 flex items-end justify-between border-b border-border pb-6">
		<div>
			<h1 class="font-sans text-headline-lg tracking-tighter">{$t('account.title')}</h1>
			{#if $customer}
				<p class="mt-2 font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted">
					{$t('account.greeting', { name: $customer.name || $customer.email })}
				</p>
			{/if}
		</div>
		<button onclick={signOut} class="btn-ghost">{$t('account.logout')}</button>
	</header>

	<h2 class="mb-6 font-sans text-headline-md tracking-tight">{$t('account.ordersTitle')}</h2>

	{#if loading}
		<p class="font-mono text-metadata-sm uppercase tracking-widest text-text-muted">{$t('common.loading')}…</p>
	{:else if failed}
		<p class="font-mono text-metadata-sm uppercase tracking-widest text-text-muted">{$t('account.ordersFailed')}</p>
	{:else if orders.length === 0}
		<div class="border border-border py-16 text-center">
			<p class="font-mono text-metadata-sm uppercase tracking-widest text-text-muted">{$t('account.ordersEmpty')}</p>
		</div>
	{:else}
		<ul class="flex flex-col gap-px bg-border">
			{#each orders as order (order.id)}
				<li class="flex flex-col gap-3 bg-bg p-5 md:flex-row md:items-start md:justify-between">
					<div class="flex flex-col gap-2">
						<span class="font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted">
							{$t('account.orderNumber')} {order.id.slice(-8)} · {fmtDate(order.createdAt)}
						</span>
						<ul class="flex flex-col gap-0.5">
							{#each order.items as item (item.sku)}
								<li class="font-sans text-body-md text-text">{item.name} ×{item.qty}</li>
							{/each}
						</ul>
					</div>
					<div class="flex items-center gap-4 md:flex-col md:items-end md:gap-2">
						<span
							class="brutal-border px-2 py-0.5 font-mono text-metadata-sm uppercase tracking-[0.05em] {order.status ===
							'CONFIRMED'
								? 'text-accent'
								: order.status === 'FAILED'
									? 'text-error'
									: 'text-text-muted'}"
						>
							{statusLabel(order)}
						</span>
						<span class="font-sans text-body-md text-text">{money(order.totalCents, order.currency, $locale)}</span>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>
