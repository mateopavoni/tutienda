import type { Handle } from '@sveltejs/kit';

// Baseline security headers for every response the storefront/dashboard serves (landing, /store/[slug],
// /app, /admin). The gateway (services/internal/platform/httpx) sets the equivalent set for the API; this
// is the same checklist for the HTML side. See ../../.claude modules/security-owasp.md §4.
export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);
	const h = response.headers;

	// No one else may frame this app (clickjacking). CSP frame-ancestors is the modern equivalent; kept
	// alongside X-Frame-Options for browsers that don't read CSP.
	h.set('X-Frame-Options', 'DENY');
	h.set('Content-Security-Policy', "frame-ancestors 'none'");
	// Stop MIME-sniffing a response into something executable.
	h.set('X-Content-Type-Options', 'nosniff');
	// Never leak the full URL (store slugs, product ids) to a third-party referrer target.
	h.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	// This app uses none of these browser features; deny them explicitly.
	h.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');
	// Only meaningful once TLS terminates in front (Nginx/Dokku in production); harmless over local http.
	if (event.url.protocol === 'https:') {
		h.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains');
	}

	return response;
};
