// Single source of truth for the SaaS brand name. Change PUBLIC_BRAND_NAME (env) to rebrand the whole
// platform — marketing landing, auth pages, merchant dashboard chrome — in one place. The per-store
// storefront uses the *store's* displayName instead; this constant is only the platform brand.
import { env } from '$env/dynamic/public';

export const BRAND = env.PUBLIC_BRAND_NAME || 'TuTienda';

// Public root domain stores live under as subdomains in production (slug.tutienda.com.ar). Used for
// copy on the landing/onboarding; the actual slug→tenant resolution happens in the gateway.
export const ROOT_DOMAIN = env.PUBLIC_ROOT_DOMAIN || 'tutienda.com.ar';
