# Decisiones

- **FACT** Atomicidad sobre locks: un `findOneAndUpdate` con guard evita locks distribuidos y transacciones multi-documento. Costo: reservas multi-SKU no son atómicas globalmente; se compensan explícitamente.
- **FACT** Saga síncrona por HTTP entre orders e inventory: simple de leer/depurar, a costa de acoplamiento temporal.
- **FACT** Monorepo de servicios (un módulo Go, `internal/platform` compartido) en vez de repos separados.
- **FACT** Multi-tenancy por colecciones compartidas + `tenantId` (modelo Shopify), no por base física.
- **FACT** Las escrituras de catálogo/stock pasan solo por `/api/admin/*` con token de tienda; la ruta pública es de solo lectura porque los backends confían en el tenant que estampa el gateway.
- **FACT** En `ENV=prod`, accounts y gateway no arrancan con `JWT_SECRET` ausente o igual al default de dev.
- **FACT** Al archivar: puertos publicados solo en `127.0.0.1`, se sacaron GA4 y la verificación de Search Console del HTML, y se eliminó el workflow de deploy.
