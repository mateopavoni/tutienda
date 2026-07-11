<script lang="ts">
	import '../app.css';
	// Importing the theme store for its side effect: on the browser it subscribes applyTheme, so the
	// saved light/dark/system preference is honoured on every page (landing, auth, storefront, dashboard).
	import '$lib/stores/theme';
	import { navigating } from '$app/stores';

	let { children } = $props();
</script>

<!-- Progressive loading: a thin top bar while SvelteKit resolves the next route's SSR load (server data
     fetch, e.g. a store's product grid). Every route gets this for free from the root layout. -->
{#if $navigating}
	<div class="nav-progress-bar" role="status" aria-label="Loading"><span></span></div>
{/if}

<!-- Root layout is intentionally chrome-less: each section brings its own. The marketing site (landing,
     auth, settings) renders bare; the storefront adds Nav/Footer/Cart under /store/[slug]; the merchant
     dashboard adds its own header under /app. -->
{@render children()}
