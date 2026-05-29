<script lang="ts">
	import '../app.css';
	// Importing the theme store for its side effect: on the browser it subscribes applyTheme, so the
	// saved light/dark/system preference is honoured on every page (landing, auth, storefront, dashboard).
	import '$lib/stores/theme';

	let { children } = $props();
</script>

<!-- Root layout is intentionally chrome-less: each section brings its own. The marketing site (landing,
     auth, settings) renders bare; the storefront adds Nav/Footer/Cart under /store/[slug]; the merchant
     dashboard adds its own header under /app. -->
{@render children()}
