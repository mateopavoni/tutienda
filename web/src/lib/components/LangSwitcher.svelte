<script lang="ts">
	import { locale, locales } from '$lib/i18n';
</script>

<div class="flex items-center gap-1 font-mono text-metadata-sm">
	{#each locales as option, i (option.code)}
		{#if i > 0}<span class="text-outline" aria-hidden="true">/</span>{/if}
		<button
			onclick={() => locale.set(option.code)}
			class="px-1 uppercase transition-colors hover:text-accent {$locale === option.code
				? 'text-text font-bold'
				: 'text-text-muted'}"
			aria-pressed={$locale === option.code}
		>
			{option.label}
		</button>
	{/each}
</div>
