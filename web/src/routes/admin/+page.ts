// The super-admin console is a pure client-side app: it reads the merchant token from localStorage and
// talks to /api/platform. Disable SSR so there's no server render without the token.
export const ssr = false;
