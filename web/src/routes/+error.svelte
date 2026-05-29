<script lang="ts">
	import { page } from '$app/stores';
	import { t } from '$lib/i18n';
	import { ArrowLeft } from 'lucide-svelte';
</script>

<section class="mx-auto flex min-h-screen max-w-container flex-col items-center justify-center gap-6 bg-bg px-4 py-32 text-center text-text">
	<h1 class="font-sans leading-none tracking-tighter text-display-xl text-text">{$page.status}</h1>
	<p class="font-mono text-metadata-sm uppercase tracking-widest text-text-muted">
		{$page.error?.message ?? $t('product.notFound')}
	</p>
	<a href="/" class="flex items-center gap-2 btn-ghost">
		<ArrowLeft size={16} />
		{$t('product.back')}
	</a>
</section>
