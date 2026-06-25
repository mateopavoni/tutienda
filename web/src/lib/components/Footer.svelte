<script lang="ts">
	import { t } from '$lib/i18n';
	import { BRAND } from '$lib/brand';
	import { pageTitle, type Page } from '$lib/pages';
	let { storeName, slug = '', pages = [] }: { storeName?: string; slug?: string; pages?: Page[] } = $props();
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
		<!-- Store pages (only ones with content) + the platform attribution; all of these actually navigate. -->
		<div class="flex flex-wrap gap-8 font-mono text-metadata-sm md:col-span-6 md:justify-end">
			{#each pages as p (p.type)}
				<a href="/store/{slug}/{p.type}" class="text-text-muted transition-colors hover:text-accent">
					{pageTitle(p)}
				</a>
			{/each}
			<a href="/" class="text-text-muted transition-colors hover:text-accent">
				{$t('footer.poweredBy', { brand: BRAND })}
			</a>
		</div>
	</div>
</footer>
