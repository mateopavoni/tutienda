<script lang="ts">
	// Shared password field: the same masked input every auth form already used, plus a show/hide toggle
	// so someone can check what they typed before submitting (a mistyped password they can't see is the
	// most common reason a login fails). One component for every form — login, signup and their
	// storefront-customer counterparts — so the affordance can't drift between them.
	//
	// The type is swapped on the *same* input element (not two inputs behind an {#if}) to keep focus and
	// the caret where they were. That rules out `bind:value` (Svelte forbids a dynamic `type` with a
	// two-way binding), so the binding is done by hand: value in, oninput out. Icons come from
	// lucide-svelte, already a dependency and already the icon set used everywhere else here.
	import { Eye, EyeOff } from 'lucide-svelte';
	import { t } from '$lib/i18n';

	let {
		value = $bindable(''),
		required = false,
		minlength,
		autocomplete = 'current-password',
		class: className = ''
	}: {
		value?: string;
		required?: boolean;
		minlength?: number;
		/** 'current-password' when signing in, 'new-password' when creating/changing one. */
		autocomplete?: 'current-password' | 'new-password';
		/** The caller's field styling (the shared `inputClass`); applied to the input itself. */
		class?: string;
	} = $props();

	let visible = $state(false);
</script>

<div class="relative flex min-w-0 items-center">
	<input
		type={visible ? 'text' : 'password'}
		{value}
		oninput={(e) => (value = e.currentTarget.value)}
		{required}
		{minlength}
		{autocomplete}
		class="w-full pr-12 {className}"
	/>
	<button
		type="button"
		onclick={() => (visible = !visible)}
		aria-label={visible ? $t('common.hidePassword') : $t('common.showPassword')}
		aria-pressed={visible}
		title={visible ? $t('common.hidePassword') : $t('common.showPassword')}
		class="focus-ring absolute right-2 p-2 text-text-muted transition-colors hover:text-accent"
	>
		{#if visible}
			<EyeOff size={18} strokeWidth={1.5} />
		{:else}
			<Eye size={18} strokeWidth={1.5} />
		{/if}
	</button>
</div>
