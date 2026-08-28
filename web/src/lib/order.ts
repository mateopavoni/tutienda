import type { Order } from '$lib/types';

// Simulated fulfillment timeline. This is a portfolio project — there's no real logistics. A paid order
// advances through these stages purely as a function of time elapsed since checkout, so the buyer can
// watch it progress on the confirmation screen and in their order history.
// derived client-side from createdAt; no backend job, no cron, no persisted fulfillment state.
export const FULFILLMENT_STAGES = ['paid', 'preparing', 'shipped', 'delivered'] as const;
export type FulfillmentStage = (typeof FULFILLMENT_STAGES)[number];

// Seconds after checkout at which each stage begins. Kept short so the simulation is watchable in a demo.
const STAGE_AT_SECONDS = [0, 8, 25, 50];

// fulfillmentIndex returns how far a paid order has progressed (0..stages-1). Returns -1 for orders that
// never shipped (failed/declined payment) — those have no fulfillment timeline.
export function fulfillmentIndex(order: Order, now: number = Date.now()): number {
	if (order.status !== 'CONFIRMED') return -1;
	const elapsed = (now - new Date(order.createdAt).getTime()) / 1000;
	let idx = 0;
	for (let i = 0; i < STAGE_AT_SECONDS.length; i++) {
		if (elapsed >= STAGE_AT_SECONDS[i]) idx = i;
	}
	return idx;
}

export function fulfillmentStage(order: Order, now: number = Date.now()): FulfillmentStage | null {
	const idx = fulfillmentIndex(order, now);
	return idx < 0 ? null : FULFILLMENT_STAGES[idx];
}
