<script lang="ts">
	import { t } from '$lib/i18n';
	import { BRAND } from '$lib/brand';
	let { storeName }: { storeName?: string } = $props();
	const year = new Date().getFullYear();
	const name = $derived(storeName ?? '');
</script>

<footer class="w-full border-t border-border bg-bg">
	<div class="mx-auto grid max-w-container grid-cols-1 gap-8 px-4 py-12 md:grid-cols-12 md:px-12">
		<div class="md:col-span-6">
			<p class="font-sans text-label-caps uppercase tracking-[0.1em] text-text">
				© {year} {name}
			</p>
		</div>
		<!-- No dead links: just the platform attribution, which actually navigates. -->
		<div class="flex flex-wrap gap-8 font-mono text-metadata-sm md:col-span-6 md:justify-end">
			<a href="/" class="text-text-muted transition-colors hover:text-accent">
				{$t('footer.poweredBy', { brand: BRAND })}
			</a>
		</div>
	</div>
</footer>
