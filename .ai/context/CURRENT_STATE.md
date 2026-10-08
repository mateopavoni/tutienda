# Estado actual (2026-10-08)

- **FACT** `gitleaks` sobre 178 commits: sin hallazgos.
- **FACT** `go vet` limpio; `go test -race` en accounts, catalog, inventory, authx y plan pasan (gateway y orders no tienen tests).
- **FACT** Web: `svelte-check` 0 errores (1 warning de tipos `node`), vitest 22/22, build ok.
- **FACT** Stack real (compose): `/health` reporta los 4 backends `ok`; web 200 en `/`, `/signup`, `/login`, `/store/system-archive`, `/app`, `/admin`, `/configuracion`, `/sitemap.xml`, `/robots.txt`.
- **FACT** Concurrencia: 20 compras simultáneas con stock 5 → ninguna sobreventa; `loadtest` (500 compradores, 10 unidades) → 10 ventas, 490 "sin stock".
- **FACT** Aislamiento: un comerciante B recibe 404 al pedir token, editar, cambiar plan/estado, leer mensajes o borrar una tienda de A; clientes y comerciantes ajenos no leen pedidos de otro; rutas `/api/platform` rechazan merchants y tokens de tienda; token inválido y `alg:none` dan 401; `X-Tenant-ID` falsificado en rutas admin da 401; escritura pública a catálogo da 405.
- **FACT** Con `ENV=prod` y el `JWT_SECRET` de dev, accounts se niega a arrancar.
- **FACT** Bug corregido al archivar: `cmd/loadtest` mandaba checkouts sin dirección de envío (todos 400) y aun así imprimía "OK ... zero oversell". Ahora envía dirección y falla si hay requests con error o si no vende exactamente el stock.
- **FACT** Playwright / navegador real no se ejecutó en esta pasada (la web se verificó por HTTP).
