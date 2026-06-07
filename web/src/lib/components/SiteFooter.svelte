<script lang="ts">
	// Marketing/platform footer. The storefront keeps its own Footer (per-store copy).
	import { BRAND } from '$lib/brand';
	import { t } from '$lib/i18n';
	const year = new Date().getFullYear();
</script>

<footer class="w-full border-t border-border bg-bg">
	<div class="mx-auto grid max-w-container grid-cols-1 gap-8 px-4 py-12 md:grid-cols-12 md:px-12">
		<div class="md:col-span-6">
			<p class="font-sans text-headline-md tracking-tighter text-text">{BRAND}</p>
			<p class="mt-2 font-mono text-metadata-sm uppercase tracking-[0.1em] text-text-muted">
				© {year} {BRAND}. {$t('landing.footerNote')}
			</p>
		</div>
		<nav class="flex flex-wrap gap-8 font-mono text-metadata-sm md:col-span-6 md:justify-end">
			<a href="/signup" class="text-text-muted transition-colors hover:text-accent">{$t('landing.cta')}</a>
			<a href="/login" class="text-text-muted transition-colors hover:text-accent">{$t('auth.login')}</a>
			<a href="/configuracion" class="text-text-muted transition-colors hover:text-accent">{$t('settings.title')}</a>
		</nav>
	</div>
</footer>
