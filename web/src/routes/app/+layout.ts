// The merchant dashboard is a pure client-side app: it holds tokens in localStorage and talks to the
// gateway directly, so there is nothing to server-render.
export const ssr = false;
export const prerender = false;
