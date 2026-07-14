<script lang="ts">
	// Platform super-admin console (cross-tenant). A pure client-side app: log in with an admin merchant,
	// then list every store/merchant, see platform stats and change any store's plan. The gateway gates
	// /api/platform on the admin role, so a non-admin token gets a 403 and is bounced back to login here.
	import { onMount } from 'svelte';
	import { BRAND } from '$lib/brand';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import { session } from '$lib/admin/session';
	import {
		login,
		platformStats,
		platformStores,
		platformMerchants,
		platformSetPlan,
		type Store,
		type Merchant,
		type PlatformStats
	} from '$lib/admin/api';
	import { PLANS, TIERS, normalizeTier } from '$lib/plan';
	import type { ApiError } from '$lib/types';

	let authed = $state(false);
	let error = $state('');
	let notice = $state('');
	// True until the onMount session check settles. Without this an already-logged-in admin sees the
	// login form flash before the stored token is verified and the console data arrives.
	let checkingSession = $state(true);

	// Login form
	let email = $state('');
	let password = $state('');

	let stats = $state<PlatformStats | null>(null);
	let stores = $state<Store[]>([]);
	let merchants = $state<Merchant[]>([]);

	const emailByOwner = $derived(new Map(merchants.map((m) => [m.id, m.email])));

	function fail(err: unknown) {
		error = (err as ApiError)?.error ?? 'Request failed';
		notice = '';
	}

	onMount(async () => {
		try {
			if (session.merchantToken() && session.isAdmin()) {
				await load();
			}
		} finally {
			checkingSession = false;
		}
	});

	async function load() {
		try {
			[stats, stores, merchants] = await Promise.all([
				platformStats(),
				platformStores(),
				platformMerchants()
			]);
			authed = true;
			error = '';
		} catch (err) {
			// A 403 means the token isn't an admin — drop back to the login form.
			authed = false;
			fail(err);
		}
	}

	async function doLogin(e: SubmitEvent) {
		e.preventDefault();
		try {
			await login(email, password);
			if (!session.isAdmin()) {
				session.clearAll();
				error = 'That account is not a platform admin.';
				return;
			}
			password = '';
			await load();
		} catch (err) {
			fail(err);
		}
	}

	function logout() {
		session.clearAll();
		authed = false;
		stats = null;
		stores = [];
		merchants = [];
	}

	async function changePlan(store: Store, plan: string) {
		try {
			const updated = await platformSetPlan(store.id, plan);
			stores = stores.map((s) => (s.id === updated.id ? updated : s));
			if (stats) {
				// Recompute the per-plan breakdown locally so the stat cards stay in sync without a refetch.
				const byPlan: Record<string, number> = {};
				for (const s of stores) byPlan[normalizeTier(s.plan)] = (byPlan[normalizeTier(s.plan)] ?? 0) + 1;
				stats = { ...stats, storesByPlan: byPlan };
			}
			notice = `${store.displayName} → ${PLANS[normalizeTier(plan)].label}`;
			error = '';
		} catch (err) {
			fail(err);
		}
	}

	const labelClass = 'font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted';
	const inputClass = 'brutal-border bg-surface px-4 py-3 font-sans text-body-md text-text focus-ring';
</script>

<svelte:head><title>{BRAND} — Platform console</title></svelte:head>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<SiteHeader variant="minimal" />

	<main class="mx-auto w-full max-w-container flex-1 px-4 py-12 md:px-12">
		<div class="mb-8 flex flex-wrap items-center justify-between gap-4">
			<div>
				<p class="font-mono text-metadata-sm uppercase tracking-[0.2em] text-accent">Super-admin</p>
				<h1 class="font-sans uppercase leading-none tracking-tighter text-headline-lg">Platform console</h1>
			</div>
			{#if authed}
				<button onclick={logout} class="btn-ghost">Log out</button>
			{/if}
		</div>

		{#if error}
			<p class="mb-6 brutal-border border-error px-4 py-3 font-mono text-metadata-sm text-error">{error}</p>
		{/if}
		{#if notice}
			<p class="mb-6 brutal-border px-4 py-3 font-mono text-metadata-sm text-text-muted">{notice}</p>
		{/if}

		{#if checkingSession}
			<!-- Verifying a stored session before deciding between the login form and the console, so a
			     returning admin doesn't see the login form flash first. -->
			<div class="flex flex-col gap-4">
				<div class="skeleton-block h-24 w-full"></div>
				<div class="skeleton-block h-48 w-full"></div>
			</div>
		{:else if !authed}
			<!-- Admin login -->
			<form onsubmit={doLogin} class="flex max-w-sm flex-col gap-4">
				<p class="font-sans text-body-md text-text-muted">Sign in with a platform admin account.</p>
				<label class="flex flex-col gap-1">
					<span class={labelClass}>Email</span>
					<input bind:value={email} type="email" required class={inputClass} />
				</label>
				<label class="flex flex-col gap-1">
					<span class={labelClass}>Password</span>
					<input bind:value={password} type="password" required class={inputClass} />
				</label>
				<button type="submit" class="btn-accent">Enter console</button>
			</form>
		{:else}
			<!-- Stats -->
			{#if stats}
				<section class="mb-12 grid grid-cols-2 gap-px brutal-border bg-border md:grid-cols-4">
					<div class="flex flex-col gap-1 bg-bg p-5">
						<span class={labelClass}>Merchants</span>
						<span class="font-sans text-headline-lg tracking-tighter">{stats.merchants}</span>
					</div>
					<div class="flex flex-col gap-1 bg-bg p-5">
						<span class={labelClass}>Stores</span>
						<span class="font-sans text-headline-lg tracking-tighter">{stats.stores}</span>
					</div>
					{#each TIERS as t (t)}
						{#if t !== 'free'}
							<div class="flex flex-col gap-1 bg-bg p-5">
								<span class={labelClass}>{PLANS[t].label} stores</span>
								<span class="font-sans text-headline-lg tracking-tighter text-accent">
									{stats.storesByPlan[t] ?? 0}
								</span>
							</div>
						{/if}
					{/each}
				</section>
			{/if}

			<!-- Stores with inline plan control -->
			<section class="mb-12">
				<h2 class="mb-4 font-sans uppercase leading-none tracking-tight text-headline-md">All stores</h2>
				<div class="flex flex-col gap-px brutal-border bg-border">
					{#each stores as store (store.id)}
						<div class="flex flex-wrap items-center justify-between gap-4 bg-bg p-4">
							<div class="flex flex-col">
								<a href="/store/{store.slug}" target="_blank" rel="noopener" class="font-sans text-body-md hover:text-accent">
									{store.displayName}
								</a>
								<span class={labelClass}>
									{store.slug} · {emailByOwner.get(store.ownerId) ?? store.ownerId}
								</span>
							</div>
							<div class="flex items-center gap-1">
								{#each TIERS as t (t)}
									<button
										onclick={() => changePlan(store, t)}
										class="px-3 py-1.5 font-mono text-metadata-sm uppercase tracking-[0.05em] transition-colors
											{normalizeTier(store.plan) === t
											? 'bg-accent text-on-accent'
											: 'brutal-border text-text-muted hover:text-text'}"
									>
										{PLANS[t].label}
									</button>
								{/each}
							</div>
						</div>
					{:else}
						<p class="bg-bg p-4 font-mono text-metadata-sm text-text-muted">No stores yet.</p>
					{/each}
				</div>
			</section>

			<!-- Merchants -->
			<section class="mb-12">
				<h2 class="mb-4 font-sans uppercase leading-none tracking-tight text-headline-md">Merchants</h2>
				<div class="flex flex-col gap-px brutal-border bg-border">
					{#each merchants as m (m.id)}
						<div class="flex items-center justify-between gap-4 bg-bg p-4">
							<span class="font-sans text-body-md">{m.email}</span>
							<span class="{labelClass} {m.role === 'admin' ? 'text-accent' : ''}">{m.role}</span>
						</div>
					{/each}
				</div>
			</section>
		{/if}
	</main>
</div>
