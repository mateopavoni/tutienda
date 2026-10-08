# Arquitectura (resumen)

- **FACT** `web` (SvelteKit, :3000) → `gateway` (:8080) → `accounts` (:8084), `catalog` (:8081), `inventory` (:8082), `orders` (:8083). Un solo Dockerfile con `ARG SERVICE`.
- **FACT** El gateway hace rate limit (Redis, ventana deslizante, `RATE_LIMIT_MAX`/min por IP), CORS, resolución de tenant por `X-Tenant-Slug` y autenticación; los backends confían en el `X-Tenant-ID` que estampa el gateway.
- **FACT** Rutas: `/api/admin/{catalog,inventory}` exigen token de tienda; `/api/catalog` y `/api/inventory` son solo lectura (GET) y públicas por slug; `/api/orders` exige token de cliente; `/api/platform` solo super-admin; la ruta interna `/admin/tenant` está bloqueada en el gateway.
- **FACT** Inventario: `findOneAndUpdate` con guard `$gte` (sin locks); reservas con TTL y un janitor que devuelve el stock vencido; checkout como saga (reserve → order PENDING → confirm, con release como compensación).
- **FACT** Redis solo para rate limit y caché de catálogo; MongoDB es la fuente de verdad.
- Detalle completo: `/ARCHITECTURE.md`.
