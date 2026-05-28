import { writable } from 'svelte/store';

// cartOpen drives the cart drawer visibility across components.
export const cartOpen = writable(false);
