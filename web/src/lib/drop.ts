// Drop mode helpers. A "drop" is a product with a future release time: before it, the storefront shows
// a countdown and checkout is blocked (also enforced server-side in the orders saga); after it, the
// product goes on sale with a live, real-time stock counter on top of the no-oversell engine.

export type DropState = 'none' | 'upcoming' | 'live';

/** dropState classifies a product by its release time relative to `now` (ms epoch, injectable for ticks). */
export function dropState(dropAt: string | null | undefined, now: number = Date.now()): DropState {
	if (!dropAt) return 'none';
	const t = new Date(dropAt).getTime();
	if (Number.isNaN(t)) return 'none';
	return now < t ? 'upcoming' : 'live';
}

export interface Countdown {
	days: number;
	hours: number;
	minutes: number;
	seconds: number;
	done: boolean;
}

/** countdown returns the time remaining until `dropAt`, clamped at zero. */
export function countdown(dropAt: string, now: number = Date.now()): Countdown {
	const ms = Math.max(0, new Date(dropAt).getTime() - now);
	const totalSeconds = Math.floor(ms / 1000);
	return {
		days: Math.floor(totalSeconds / 86400),
		hours: Math.floor((totalSeconds % 86400) / 3600),
		minutes: Math.floor((totalSeconds % 3600) / 60),
		seconds: totalSeconds % 60,
		done: ms === 0
	};
}

const pad = (n: number) => n.toString().padStart(2, '0');

/** formatCountdown renders a compact, locale-neutral timer ("2d 04:12:09" or "04:12:09"). */
export function formatCountdown(c: Countdown): string {
	const hms = `${pad(c.hours)}:${pad(c.minutes)}:${pad(c.seconds)}`;
	return c.days > 0 ? `${c.days}d ${hms}` : hms;
}
