// Shared API types, mirroring the JSON shapes the gateway returns.

export interface Variant {
	sku: string;
	label: string;
	color?: string;
}

export interface Product {
	id: string;
	name: string;
	category: string;
	description: string;
	priceCents: number;
	currency: string;
	/** Primary image (grid card + cart thumbnail). Always equals images[0] when a gallery exists. */
	imageUrl: string;
	/** Full image gallery shown on the product detail page. */
	images?: string[];
	variants: Variant[];
	/** What a variant's label means for this product ("Size", "Color", "Storage"...). Blank/absent falls
	 *  back to a generic "Size" in the storefront selector. */
	variantLabel?: string;
	createdAt: string;
	/** ISO release time for a scheduled drop. Absent/null = on sale now. */
	dropAt?: string | null;
}

export interface Stock {
	sku: string;
	available: number;
	reserved: number;
}

export interface OrderItem {
	sku: string;
	qty: number;
	name: string;
	priceCents: number;
}

export interface ShippingAddress {
	name: string;
	line1: string;
	city: string;
	postalCode: string;
	country: string;
}

export interface Order {
	id: string;
	items: OrderItem[];
	shipping?: ShippingAddress;
	currency: string;
	shippingCents?: number;
	totalCents: number;
	status: 'PENDING' | 'CONFIRMED' | 'FAILED';
	paymentStatus?: 'paid' | 'failed';
	reservationId: string;
	failureReason?: string;
	createdAt: string;
}

/** A submission from a store's storefront "contact" section (see /app "Messages"). */
export interface Message {
	id: string;
	name: string;
	email: string;
	body: string;
	createdAt: string;
}

export interface ApiError {
	error: string;
	detail?: string;
	sku?: string;
	/** HTTP status of the failed response, stamped by the client — not sent by the server. */
	status?: number;
}
