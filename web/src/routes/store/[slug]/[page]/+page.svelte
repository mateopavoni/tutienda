<script lang="ts">
	import type { PageData } from './$types';
	import { pageTitle } from '$lib/pages';

	let { data }: { data: PageData } = $props();
	const page = $derived(data.page);
	const title = $derived(pageTitle(page));
	const storeName = $derived(data.store?.displayName ?? data.storeSlug);
</script>

<svelte:head>
	<title>{storeName} — {title}</title>
</svelte:head>

<section class="mx-auto w-full max-w-3xl px-4 py-16 md:px-12 md:py-24">
	<h1
		class="font-sans uppercase leading-none tracking-tighter text-headline-lg text-text"
		style="color: rgb(var(--title-color))"
	>
		{title}
	</h1>
	<!-- Body is plain text the merchant typed; whitespace-pre-line keeps their line breaks without HTML. -->
	<div class="mt-10 whitespace-pre-line font-sans text-body-md leading-relaxed text-text-muted">
		{page.body}
	</div>
</section>
