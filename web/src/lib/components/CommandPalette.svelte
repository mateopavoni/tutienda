<script lang="ts">
	import { t } from '$lib/i18n';
	import { theme } from '$lib/stores/theme';

	// Merchant-dashboard-only quick nav: Ctrl/Cmd+K opens a filterable list of fixed actions (scroll to a
	// section, focus the new-product field, toggle theme). No fuzzy search / plugin actions — the list is
	// short and fixed, a substring filter is plenty.
	let dialogEl: HTMLDialogElement;
	let query = $state('');

	interface Action {
		id: string;
		label: string;
		run: () => void;
	}

	function goTo(sectionId: string, focusId?: string) {
		return () => {
			const el = document.getElementById(sectionId);
			if (!el) return;
			el.scrollIntoView({ behavior: 'smooth', block: 'start' });
			if (focusId) document.getElementById(focusId)?.focus();
		};
	}

	function cycleTheme() {
		const next = { system: 'light', light: 'dark', dark: 'system' } as const;
		theme.update((v) => next[v]);
	}

	const actions = $derived<Action[]>([
		{ id: 'new-product', label: $t('app.palette.newProduct'), run: goTo('app-products', 'app-product-name') },
		{ id: 'go-stores', label: $t('app.palette.goStores'), run: goTo('app-stores') },
		{ id: 'go-membership', label: $t('app.palette.goMembership'), run: goTo('app-membership') },
		{ id: 'go-products', label: $t('app.palette.goProducts'), run: goTo('app-products') },
		{ id: 'go-messages', label: $t('app.palette.goMessages'), run: goTo('app-messages') },
		{ id: 'go-settings', label: $t('app.palette.goSettings'), run: goTo('app-settings') },
		{ id: 'toggle-theme', label: $t('app.palette.toggleTheme'), run: cycleTheme }
	]);

	const filtered = $derived(
		query.trim()
			? actions.filter((a) => a.label.toLowerCase().includes(query.trim().toLowerCase()))
			: actions
	);

	function open() {
		query = '';
		dialogEl?.showModal();
	}

	function run(action: Action) {
		action.run();
		dialogEl?.close();
	}

	function onKeydown(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
			e.preventDefault();
			if (dialogEl?.open) dialogEl.close();
			else open();
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<dialog
	bind:this={dialogEl}
	class="w-full max-w-lg border border-border bg-bg p-0 text-text backdrop:bg-black/50"
>
	<div class="flex flex-col">
		<!-- svelte-ignore a11y_autofocus: only fires on the Ctrl/Cmd+K the user just pressed, never on load -->
		<input
			autofocus
			bind:value={query}
			placeholder={$t('app.palette.placeholder')}
			class="border-b border-border bg-transparent px-4 py-3 font-sans text-body-md text-text focus:outline-none"
		/>
		<ul class="max-h-80 overflow-y-auto">
			{#each filtered as action (action.id)}
				<li>
					<button
						type="button"
						onclick={() => run(action)}
						class="w-full px-4 py-2 text-left font-mono text-metadata-sm uppercase tracking-[0.05em] text-text-muted hover:bg-surface hover:text-text"
					>
						{action.label}
					</button>
				</li>
			{:else}
				<li class="px-4 py-3 font-mono text-metadata-sm text-text-muted">{$t('app.palette.empty')}</li>
			{/each}
		</ul>
	</div>
</dialog>
