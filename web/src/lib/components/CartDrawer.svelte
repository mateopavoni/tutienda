<script lang="ts">
	import { cart, cartTotal, cartCurrency } from '$lib/stores/cart';
	import { cartOpen } from '$lib/stores/ui';
	import { t, locale } from '$lib/i18n';
	import { money } from '$lib/format';
	import { checkout } from '$lib/api/client';
	import type { ApiError, Order } from '$lib/types';
	import { X, ShoppingBag, Minus, Plus, ArrowRight, Check, TriangleAlert } from 'lucide-svelte';

	type Phase = 'idle' | 'processing' | 'confirmed' | 'failed';
	let phase = $state<Phase>('idle');
	let order = $state<Order | null>(null);
	let soldOut = $state(false);

	function close() {
		if (phase === 'confirmed') cart.clear();
		phase = 'idle';
		order = null;
		soldOut = false;
		cartOpen.set(false);
	}

	async function placeOrder() {
		phase = 'processing';
		soldOut = false;
		try {
			const items = $cart.map((line) => ({ sku: line.sku, qty: line.qty }));
			order = await checkout(items);
			phase = order.status === 'CONFIRMED' ? 'confirmed' : 'failed';
		} catch (e) {
			const err = e as ApiError;
			soldOut = Boolean(err?.sku);
			phase = 'failed';
		}
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && $cartOpen) close();
	}
</script>

<svelte:window on:keydown={onKeydown} />

{#if $cartOpen}
	<button
		class="fixed inset-0 z-[90] bg-black/40 backdrop-blur-sm"
		onclick={close}
		aria-label={$t('common.close')}
	></button>
{/if}

<div
	class="fixed inset-y-0 right-0 z-[100] flex w-full max-w-sm flex-col border-l border-on-inverse/20 bg-inverse text-on-inverse transition-transform duration-500 ease-in-out {$cartOpen
		? 'translate-x-0'
		: 'translate-x-full'}"
	role="dialog"
	aria-modal="true"
	aria-label={$t('cart.title')}
>
	<header class="flex items-center justify-between border-b border-on-inverse/20 p-8">
		<div>
			<h2 class="font-sans text-headline-md">{$t('cart.title')}</h2>
			<p class="mt-1 font-mono text-metadata-sm uppercase tracking-widest opacity-60">
				{$t('cart.subtitle')}
			</p>
		</div>
		<button onclick={close} class="transition-opacity hover:opacity-60" aria-label={$t('common.close')}>
			<X size={20} strokeWidth={1.5} />
		</button>
	</header>

	{#if phase === 'confirmed' && order}
		<div class="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
			<Check size={40} strokeWidth={1.5} class="text-accent" />
			<h3 class="font-sans text-headline-md">{$t('cart.confirmedTitle')}</h3>
			<p class="font-mono text-metadata-sm uppercase tracking-widest opacity-70">
				{$t('cart.orderId')}: {order.id.slice(-8)}
			</p>
			<p class="font-mono text-metadata-sm">{money(order.totalCents, order.currency, $locale)}</p>
			<button onclick={close} class="mt-4 btn-accent w-full">{$t('cart.continue')}</button>
		</div>
	{:else if phase === 'failed'}
		<div class="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
			<TriangleAlert size={40} strokeWidth={1.5} class="text-error" />
			<h3 class="font-sans text-headline-md">{$t('cart.failedTitle')}</h3>
			<p class="font-mono text-metadata-sm opacity-70">
				{soldOut ? $t('cart.soldOut') : $t('catalog.errorLoad')}
			</p>
			<button onclick={() => (phase = 'idle')} class="mt-4 btn-accent w-full">
				{$t('common.retry')}
			</button>
		</div>
	{:else if $cart.length === 0}
		<div class="flex flex-1 flex-col items-center justify-center gap-4 opacity-60">
			<ShoppingBag size={36} strokeWidth={1.5} />
			<p class="font-mono text-metadata-sm uppercase tracking-widest">{$t('cart.empty')}</p>
		</div>
	{:else}
		<div class="flex-1 overflow-y-auto">
			{#each $cart as line (line.sku)}
				<div class="flex gap-5 border-b border-on-inverse/20 p-6">
					<div class="h-28 w-20 shrink-0 border border-on-inverse/20 bg-white/5">
						<img src={line.imageUrl} alt={line.name} class="h-full w-full object-cover grayscale" />
					</div>
					<div class="flex flex-1 flex-col justify-between">
						<div>
							<h3 class="font-mono text-metadata-sm uppercase tracking-widest">{line.name}</h3>
							<p class="font-mono text-metadata-sm opacity-60">{line.label}</p>
						</div>
						<div class="flex items-end justify-between">
							<div class="flex items-center border border-on-inverse/30">
								<button
									onclick={() => cart.setQty(line.sku, line.qty - 1)}
									class="flex h-8 w-8 items-center justify-center transition-colors hover:bg-on-inverse hover:text-inverse"
									aria-label={$t('common.remove')}
								>
									<Minus size={14} />
								</button>
								<span class="w-8 text-center font-mono text-metadata-sm">{line.qty}</span>
								<button
									onclick={() => cart.setQty(line.sku, line.qty + 1)}
									class="flex h-8 w-8 items-center justify-center transition-colors hover:bg-on-inverse hover:text-inverse"
									aria-label={$t('common.quantity')}
								>
									<Plus size={14} />
								</button>
							</div>
							<span class="font-sans text-headline-md tracking-tighter">
								{money(line.priceCents * line.qty, line.currency, $locale)}
							</span>
						</div>
					</div>
				</div>
			{/each}
		</div>

		<div class="border-t border-on-inverse/20 p-8">
			<div class="mb-6 flex items-end justify-between">
				<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">
					{$t('common.subtotal')}
				</span>
				<span class="font-sans text-headline-md tracking-tighter">
					{money($cartTotal, $cartCurrency, $locale)}
				</span>
			</div>
			<button
				onclick={placeOrder}
				disabled={phase === 'processing'}
				class="flex w-full items-center justify-center gap-2 btn-accent disabled:opacity-60"
			>
				{phase === 'processing' ? $t('cart.processing') : $t('cart.checkoutNow')}
				<ArrowRight size={16} />
			</button>
		</div>
	{/if}
</div>
