// See https://kit.svelte.dev/docs/types#app
declare global {
	namespace App {
		interface Error {
			message: string;
		}
		// eslint-disable-next-line @typescript-eslint/no-empty-interface
		interface Locals {}
		// eslint-disable-next-line @typescript-eslint/no-empty-interface
		interface PageData {}
		// eslint-disable-next-line @typescript-eslint/no-empty-interface
		interface Platform {}
	}
}

export {};
