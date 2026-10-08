# Proyecto

- **FACT** SaaS de ecommerce multi-tienda (estilo Tienda Nube/Shopify): cada comerciante crea sus tiendas y productos; el problema central es no sobrevender stock bajo carga concurrente, aislado por tienda.
- **FACT** Stack: Go 1.22 (un módulo en `services/`, 5 binarios: gateway, accounts, catalog, inventory, orders), SvelteKit 5 (`web/`), MongoDB 7, Redis 7, MinIO (imágenes).
- **FACT** Estado: **archivado** (2026-10-08). El deploy (Dokku para gateway+web, `docker compose` en un VPS para el resto) fue dado de baja; se corre en local con `docker compose up --build`.
- **FACT** Autor: Mateo Pavoni (`mateopavoni`). Licencia propietaria, solo evaluación/portfolio.
- **INFERENCE** Pieza de portfolio: el valor está en la garantía de inventario, la saga de checkout y el aislamiento por tenant, no en el CRUD.
