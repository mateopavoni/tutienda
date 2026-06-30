<script lang="ts">
	import { cart, cartTotal, cartCurrency } from '$lib/stores/cart';
	import { cartOpen } from '$lib/stores/ui';
	import { t, locale } from '$lib/i18n';
	import { money } from '$lib/format';
	import { checkout } from '$lib/api/client';
	import type { ApiError, Order, ShippingAddress } from '$lib/types';
	import { FULFILLMENT_STAGES, fulfillmentIndex } from '$lib/order';
	import { X, ShoppingBag, Minus, Plus, ArrowRight, ArrowLeft, Check, TriangleAlert, CreditCard } from 'lucide-svelte';

	// ponytail: mirror of orders.flatShippingCents (Go is source of truth; the server total is authoritative).
	const SHIPPING_FLAT_CENTS = 1500;

	type Phase = 'idle' | 'shipping' | 'payment' | 'processing' | 'confirmed' | 'failed';
	let phase = $state<Phase>('idle');
	let order = $state<Order | null>(null);
	let soldOut = $state(false);
	let declined = $state(false);

	let shipping = $state<ShippingAddress>({ name: '', line1: '', city: '', postalCode: '', country: '' });
	let approve = $state(true); // simulated payment decision

	const shippingValid = $derived(
		shipping.name.trim() !== '' &&
			shipping.line1.trim() !== '' &&
			shipping.city.trim() !== '' &&
			shipping.country.trim() !== ''
	);

	function close() {
		if (phase === 'confirmed') cart.clear();
		phase = 'idle';
		order = null;
		soldOut = false;
		declined = false;
		approve = true;
		cartOpen.set(false);
	}

	async function placeOrder() {
		phase = 'processing';
		soldOut = false;
		declined = false;
		try {
			const items = $cart.map((line) => ({ sku: line.sku, qty: line.qty }));
			order = await checkout(items, shipping, approve);
			if (order.status === 'CONFIRMED') {
				phase = 'confirmed';
			} else {
				declined = order.paymentStatus === 'failed';
				phase = 'failed';
			}
		} catch (e) {
			const err = e as ApiError;
			soldOut = Boolean(err?.sku);
			phase = 'failed';
		}
	}

	// Tick a clock while a confirmed order is on screen so its simulated fulfillment stage advances live.
	let now = $state(Date.now());
	$effect(() => {
		if (phase !== 'confirmed') return;
		const id = setInterval(() => (now = Date.now()), 1000);
		return () => clearInterval(id);
	});
	const stageIndex = $derived(order ? fulfillmentIndex(order, now) : -1);

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
		<div class="flex flex-1 flex-col gap-6 overflow-y-auto p-8">
			<div class="flex flex-col items-center gap-3 text-center">
				<Check size={40} strokeWidth={1.5} class="text-accent" />
				<h3 class="font-sans text-headline-md">{$t('cart.confirmedTitle')}</h3>
				<p class="font-mono text-metadata-sm uppercase tracking-widest opacity-70">
					{$t('cart.orderId')}: {order.id.slice(-8)}
				</p>
				<p class="font-mono text-metadata-sm">{money(order.totalCents, order.currency, $locale)}</p>
			</div>

			<!-- Simulated fulfillment tracker: advances on its own while this screen is open. -->
			<div class="border-t border-on-inverse/20 pt-6">
				<p class="mb-4 font-mono text-metadata-sm uppercase tracking-widest opacity-60">
					{$t('cart.trackingTitle')}
				</p>
				<ol class="flex flex-col gap-3">
					{#each FULFILLMENT_STAGES as stage, i (stage)}
						<li class="flex items-center gap-3">
							<span
								class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-[11px] transition-colors {i <=
								stageIndex
									? 'border-accent bg-accent text-on-accent'
									: 'border-on-inverse/30 opacity-50'}"
							>
								{#if i < stageIndex}<Check size={12} />{:else}{i + 1}{/if}
							</span>
							<span
								class="font-mono text-metadata-sm uppercase tracking-widest {i === stageIndex
									? 'text-accent'
									: i < stageIndex
										? ''
										: 'opacity-50'}"
							>
								{$t('cart.stage_' + stage)}
							</span>
						</li>
					{/each}
				</ol>
			</div>

			<button onclick={close} class="mt-auto btn-accent w-full">{$t('cart.continue')}</button>
		</div>
	{:else if phase === 'failed'}
		<div class="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
			<TriangleAlert size={40} strokeWidth={1.5} class="text-error" />
			<h3 class="font-sans text-headline-md">{$t('cart.failedTitle')}</h3>
			<p class="font-mono text-metadata-sm opacity-70">
				{soldOut ? $t('cart.soldOut') : declined ? $t('cart.declined') : $t('catalog.errorLoad')}
			</p>
			<button onclick={() => (phase = declined ? 'payment' : 'idle')} class="mt-4 btn-accent w-full">
				{$t('common.retry')}
			</button>
		</div>
	{:else if phase === 'shipping'}
		<div class="flex flex-1 flex-col overflow-y-auto p-8">
			<button onclick={() => (phase = 'idle')} class="mb-4 flex items-center gap-1 self-start font-mono text-metadata-sm uppercase tracking-widest opacity-60 transition-opacity hover:opacity-100">
				<ArrowLeft size={14} /> {$t('common.back')}
			</button>
			<h3 class="mb-6 font-sans text-headline-md">{$t('cart.shippingTitle')}</h3>
			<div class="flex flex-col gap-4">
				<label class="flex flex-col gap-1">
					<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.fieldName')}</span>
					<input bind:value={shipping.name} class="checkout-input" autocomplete="name" />
				</label>
				<label class="flex flex-col gap-1">
					<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.fieldAddress')}</span>
					<input bind:value={shipping.line1} class="checkout-input" autocomplete="street-address" />
				</label>
				<div class="flex gap-4">
					<label class="flex flex-1 flex-col gap-1">
						<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.fieldCity')}</span>
						<input bind:value={shipping.city} class="checkout-input" autocomplete="address-level2" />
					</label>
					<label class="flex w-28 flex-col gap-1">
						<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.fieldPostal')}</span>
						<input bind:value={shipping.postalCode} class="checkout-input" autocomplete="postal-code" />
					</label>
				</div>
				<label class="flex flex-col gap-1">
					<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.fieldCountry')}</span>
					<input bind:value={shipping.country} class="checkout-input" autocomplete="country-name" />
				</label>
			</div>
			<button
				onclick={() => (phase = 'payment')}
				disabled={!shippingValid}
				class="mt-8 flex w-full items-center justify-center gap-2 btn-accent disabled:opacity-40"
			>
				{$t('cart.continueToPayment')} <ArrowRight size={16} />
			</button>
		</div>
	{:else if phase === 'payment'}
		<div class="flex flex-1 flex-col overflow-y-auto p-8">
			<button onclick={() => (phase = 'shipping')} class="mb-4 flex items-center gap-1 self-start font-mono text-metadata-sm uppercase tracking-widest opacity-60 transition-opacity hover:opacity-100">
				<ArrowLeft size={14} /> {$t('common.back')}
			</button>
			<h3 class="mb-2 font-sans text-headline-md">{$t('cart.paymentTitle')}</h3>
			<p class="mb-6 font-mono text-metadata-sm opacity-60">{$t('cart.demoNote')}</p>

			<!-- Decorative card form — nothing here is sent anywhere; the toggle below simulates the result. -->
			<div class="flex flex-col gap-4">
				<label class="flex flex-col gap-1">
					<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.cardNumber')}</span>
					<input value="4242 4242 4242 4242" readonly class="checkout-input opacity-70" />
				</label>
				<div class="flex gap-4">
					<label class="flex flex-1 flex-col gap-1">
						<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.cardExpiry')}</span>
						<input value="12/30" readonly class="checkout-input opacity-70" />
					</label>
					<label class="flex w-24 flex-col gap-1">
						<span class="font-mono text-metadata-sm uppercase tracking-widest opacity-60">{$t('cart.cardCvc')}</span>
						<input value="123" readonly class="checkout-input opacity-70" />
					</label>
				</div>
			</div>

			<div class="mt-6 flex gap-px overflow-hidden rounded-[var(--radius-control)] border border-on-inverse/30">
				<button
					onclick={() => (approve = true)}
					class="flex-1 py-2 font-mono text-metadata-sm uppercase tracking-widest transition-colors {approve ? 'bg-accent text-on-accent' : 'opacity-60'}"
				>
					{$t('cart.approve')}
				</button>
				<button
					onclick={() => (approve = false)}
					class="flex-1 py-2 font-mono text-metadata-sm uppercase tracking-widest transition-colors {!approve ? 'bg-error text-white' : 'opacity-60'}"
				>
					{$t('cart.decline')}
				</button>
			</div>

			<div class="mt-6 flex flex-col gap-1 border-t border-on-inverse/20 pt-4 font-mono text-metadata-sm">
				<div class="flex justify-between opacity-60">
					<span>{$t('common.subtotal')}</span>
					<span>{money($cartTotal, $cartCurrency, $locale)}</span>
				</div>
				<div class="flex justify-between opacity-60">
					<span>{$t('cart.shippingLine')}</span>
					<span>{money(SHIPPING_FLAT_CENTS, $cartCurrency, $locale)}</span>
				</div>
				<div class="mt-1 flex justify-between text-body-md">
					<span>{$t('common.total')}</span>
					<span class="font-sans tracking-tighter">{money($cartTotal + SHIPPING_FLAT_CENTS, $cartCurrency, $locale)}</span>
				</div>
			</div>

			<button
				onclick={placeOrder}
				class="mt-6 flex w-full items-center justify-center gap-2 btn-accent"
			>
				<CreditCard size={16} /> {$t('cart.payNow')}
			</button>
		</div>
	{:else if phase === 'processing'}
		<div class="flex flex-1 flex-col items-center justify-center gap-4 opacity-70">
			<CreditCard size={36} strokeWidth={1.5} class="animate-pulse" />
			<p class="font-mono text-metadata-sm uppercase tracking-widest">{$t('cart.processing')}…</p>
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
					<div class="h-28 w-20 shrink-0 overflow-hidden rounded-[var(--radius-surface)] border border-on-inverse/20 bg-white/5">
						<img src={line.imageUrl} alt={line.name} class="h-full w-full object-cover" />
					</div>
					<div class="flex flex-1 flex-col justify-between">
						<div>
							<h3 class="font-mono text-metadata-sm uppercase tracking-widest">{line.name}</h3>
							<p class="font-mono text-metadata-sm opacity-60">{line.label}</p>
						</div>
						<div class="flex items-end justify-between">
							<div class="flex items-center overflow-hidden rounded-[var(--radius-control)] border border-on-inverse/30">
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
				onclick={() => (phase = 'shipping')}
				class="flex w-full items-center justify-center gap-2 btn-accent"
			>
				{$t('cart.checkoutNow')}
				<ArrowRight size={16} />
			</button>
		</div>
	{/if}
</div>
