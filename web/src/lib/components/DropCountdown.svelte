<script lang="ts">
	import { countdown, formatCountdown } from '$lib/drop';

	// Ticking countdown to a drop's release. Calls `onlive` once when it reaches zero so the parent can
	// flip from "upcoming" to "live" without its own timer.
	let { dropAt, onlive }: { dropAt: string; onlive?: () => void } = $props();

	let now = $state(Date.now());
	let fired = false;

	$effect(() => {
		const id = setInterval(() => {
			now = Date.now();
		}, 1000);
		return () => clearInterval(id);
	});

	const c = $derived(countdown(dropAt, now));
	$effect(() => {
		if (c.done && !fired) {
			fired = true;
			onlive?.();
		}
	});
</script>

<span class="font-mono tabular-nums tracking-tight">{formatCountdown(c)}</span>
