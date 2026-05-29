<script lang="ts">
	import { theme, type Theme } from '$lib/stores/theme';
	import { t } from '$lib/i18n';
	import { Sun, Moon, Monitor } from 'lucide-svelte';

	const order: Theme[] = ['system', 'light', 'dark'];
	function cycle() {
		theme.update((current) => order[(order.indexOf(current) + 1) % order.length]);
	}

	const Icon = $derived($theme === 'light' ? Sun : $theme === 'dark' ? Moon : Monitor);
</script>

<button
	onclick={cycle}
	class="p-2 text-text hover:text-accent transition-colors"
	aria-label={$t('theme.label')}
	title={$t('theme.' + $theme)}
>
	<Icon size={18} strokeWidth={1.5} />
</button>
