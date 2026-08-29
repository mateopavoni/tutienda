import type { RequestHandler } from './$types';

const BASE = 'https://tutienda.mateopavoni.com.ar';

// Landing (SaaS) + las tiendas demo semilla (mismas que muestra src/routes/+page.svelte).
// lista estática = espejo del seed del backend. Si las demo stores dejan de ser
// fijas, reemplazar por un fetch a `${PUBLIC_API_BASE}` de tenants públicos.
const DEMO_STORES = ['system-archive', 'casa-bruta', 'papel-y-tinta', 'cafe-noventa'];

export const prerender = true;

export const GET: RequestHandler = () => {
  const urls = [
    { loc: `${BASE}/`, priority: '1.0', changefreq: 'monthly' },
    ...DEMO_STORES.map((slug) => ({
      loc: `${BASE}/store/${slug}`,
      priority: '0.8',
      changefreq: 'weekly',
    })),
  ];

  const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls
  .map((u) => `  <url>\n    <loc>${u.loc}</loc>\n    <changefreq>${u.changefreq}</changefreq>\n    <priority>${u.priority}</priority>\n  </url>`)
  .join('\n')}
</urlset>`;

  return new Response(body, {
    headers: { 'Content-Type': 'application/xml' },
  });
};
