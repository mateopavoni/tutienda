import { describe, it, expect } from 'vitest';
import { fulfillmentIndex, fulfillmentStage } from './order';
import type { Order } from './types';

function paidOrder(secondsAgo: number): Order {
	return {
		id: 'o1',
		items: [],
		currency: 'USD',
		totalCents: 0,
		status: 'CONFIRMED',
		reservationId: 'r1',
		createdAt: new Date(Date.now() - secondsAgo * 1000).toISOString()
	};
}

describe('simulated fulfillment timeline', () => {
	it('advances through stages as time elapses', () => {
		expect(fulfillmentStage(paidOrder(0))).toBe('paid');
		expect(fulfillmentStage(paidOrder(10))).toBe('preparing');
		expect(fulfillmentStage(paidOrder(30))).toBe('shipped');
		expect(fulfillmentStage(paidOrder(60))).toBe('delivered');
	});

	it('caps at delivered and never goes past', () => {
		expect(fulfillmentIndex(paidOrder(10_000))).toBe(3);
	});

	it('has no timeline for non-confirmed orders', () => {
		const failed = { ...paidOrder(60), status: 'FAILED' as const };
		expect(fulfillmentIndex(failed)).toBe(-1);
		expect(fulfillmentStage(failed)).toBeNull();
	});
});
