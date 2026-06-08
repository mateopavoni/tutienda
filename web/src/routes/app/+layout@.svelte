<script lang="ts">
	// Merchant dashboard chrome. `@`-reset detaches from the root layout so the dashboard owns its full
	// frame. Unauthenticated visitors bounce to the platform login.
	import '../../app.css';
	import '$lib/stores/theme';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { session } from '$lib/admin/session';
	import { BRAND } from '$lib/brand';

	let { children } = $props();

	onMount(() => {
		if (!session.merchantToken()) goto('/login');
	});

	function logout() {
		session.clearAll();
		goto('/login');
	}
</script>

<div class="flex min-h-screen flex-col bg-bg text-text">
	<header class="flex h-16 items-center justify-between border-b border-border px-4 md:px-12">
		<a href="/app" class="font-sans text-headline-md tracking-tighter">{BRAND} <span class="text-text-muted">/ Panel</span></a>
		<div class="flex items-center gap-4">
			<a
				href="/configuracion"
				class="font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted hover:text-text"
			>
				Configuración
			</a>
			<button
				onclick={logout}
				class="font-sans text-label-caps uppercase tracking-[0.1em] text-text-muted hover:text-error"
			>
				Log out
			</button>
		</div>
	</header>
	<main class="mx-auto w-full max-w-container flex-1 px-4 py-10 md:px-12">
		{@render children()}
	</main>
</div>
