# Architecture — archive-commerce

A multi-store ecommerce builder (multi-tenant SaaS) designed around one hard problem: **never oversell
inventory under concurrent load**, now enforced **per store**. Merchants register, each builds their own
storefront under an isolated `tenantId`. Everything else (accounts, catalog, orders, gateway) exists to
make that guarantee usable end to end across many stores.

The SvelteKit front is a real SaaS, not a single storefront: `/` is the marketing landing, `/signup`
onboards a merchant (account + first store), each store's storefront lives at `/store/[slug]` (a
subdomain in production), the merchant dashboard is `/app`, and `/admin` is reserved for the platform
super-admin. Preferences (language/theme) live in `/configuracion`.

## Components

```
                       ┌─────────────────────────┐
   Browser  ──────────►│  web (SvelteKit, :3000)  │  landing · /store/[slug] · /app · /login
                       └────────────┬────────────┘
                                    │ HTTP/JSON
                       ┌────────────▼────────────┐
                       │   gateway (Go, :8080)    │  rate limit (Redis), CORS, routing
                       └───┬───────────┬──────────┘
             /api/catalog  │           │  /api/orders
                  ┌────────▼──┐   ┌─────▼──────┐      ┌──────────────┐
                  │  catalog  │   │   orders   │─────►│  inventory   │
                  │ (Go :8081)│   │ (Go :8083) │ HTTP │ (Go :8082)   │
                  └─────┬─────┘   └─────┬──────┘      └──────┬───────┘
                        │               │                    │
                  ┌─────▼─────┐   ┌──────▼──────┐      ┌──────▼───────┐
                  │  Redis    │   │   MongoDB   │      │   MongoDB    │
                  │ (cache)   │   │ orders DB   │      │ inventory DB │
                  └───────────┘   └─────────────┘      └──────────────┘
```

All four Go services are built from a **single Go module** (`services/`) with one `cmd/<service>`
entrypoint each. They share `internal/platform` (config, Mongo/Redis clients, HTTP middleware,
structured logging). Each is still an independently deployable Docker image (one `Dockerfile`,
build `ARG SERVICE`). This is a pragmatic "modular monolith of services": microservice boundaries
and independent deploys, without duplicating infrastructure code across four repos.

### gateway (:8080)
Single public entry point. Responsibilities:
- **Routing** — `/api/catalog/*`, `/api/inventory/*`, `/api/orders/*` reverse-proxied to the right service.
- **Rate limiting** — Redis sliding-window counter per client IP (`RATE_LIMIT_MAX` per `RATE_LIMIT_WINDOW_SECONDS`). Distributed: works the same with N gateway replicas because the counter lives in Redis.
- **CORS** — only the storefront origin (`WEB_ORIGIN`).
- **Health aggregation** — `/health` fans out to each service and reports per-service status.

### catalog (:8081)
Read-heavy product service. Products live as MongoDB documents (variants, images, price as integer
cents). Listing and detail are served **cache-aside** through Redis: read Redis → on miss read Mongo
→ populate Redis with a TTL. Writes (seed) invalidate the relevant keys.

### inventory (:8082) — the concurrency core
Owns stock per SKU and short-lived **reservations**.

- Stock document: `{ sku, available, reserved }`.
- **Reserve** (`POST /reservations`): for each line, atomically
  `findOneAndUpdate({sku, available: {$gte: qty}}, {$inc: {available: -qty, reserved: +qty}})`.
  The `$gte` guard makes the update match **zero documents** when stock is insufficient — the operation
  fails cleanly instead of going negative. No application-level lock, no distributed transaction.
  If a later line in the same request fails, the already-reserved lines are **compensated** (released).
- **Confirm** (`POST /reservations/{id}/confirm`): the goods leave for good — decrement `reserved`,
  mark the reservation `CONFIRMED`.
- **Release** (`POST /reservations/{id}/release`): give stock back — `reserved → available`.
- **Janitor** (background goroutine): a TTL index alone would only *delete* an expired reservation
  document and silently leak the held stock. Instead the janitor periodically finds reservations past
  `expiresAt`, **compensates** their stock back to `available`, then removes them. Expiry actually
  frees inventory.

### orders (:8083)
Checkout orchestration as a **saga** (no external orchestrator):
1. `reserve` against inventory — fail fast with `409` if any line is out of stock.
2. create the order document (`PENDING`) holding the `reservationId`.
3. `confirm` the reservation (simulated payment success) → order `CONFIRMED`.
4. on any failure after step 1 → `release` the reservation (compensating action) → order `FAILED`.

## Data flow — a checkout under contention
1. Browser POSTs a cart to `gateway /api/orders`.
2. Gateway rate-limits, forwards to `orders`.
3. `orders` calls `inventory /reservations`. Inventory runs the atomic guarded update per line.
   When 1 000 requests hit the last 10 units, exactly 10 reservations succeed; the rest get `409`.
4. Successful reservations become `PENDING` orders, then `CONFIRMED` after payment.
5. Abandoned reservations expire and the janitor returns their stock.

The invariant — `available` never goes below 0, and `available + reserved + sold` is conserved — is
asserted by `services/internal/inventory/service_test.go`, which fires hundreds of goroutines at a limited SKU.

## Key decisions and trade-offs
- **Atomicity over locking.** Document-level atomic updates in MongoDB are enough for the single-SKU
  oversell problem and avoid the cost/complexity of distributed locks or multi-document transactions.
  Trade-off: a multi-SKU reservation is not globally atomic — handled with explicit compensation.
- **Synchronous saga.** HTTP calls between orders and inventory keep the flow easy to read and debug,
  at the cost of temporal coupling. An event-driven version (outbox + queue) is the natural next step
  and is called out as a known limitation, not hidden.
- **Reservations + janitor over hard holds.** Carts get a real hold on stock, but abandoned carts can't
  leak inventory forever. The janitor makes expiry a first-class, observable action.
- **Monorepo of services.** One module, shared `platform`, four images. Less duplication than four repos,
  still independently deployable.
- **Redis as a secondary layer**, never the source of truth: rate limiting + catalog cache only.

## Styling
Design system from a Stitch AI export — "Refined Brutalism": monochrome, flat, zero radius, 1px
borders, Inter + JetBrains Mono. Tokens documented in `docs/design/design-tokens.md` and mapped to
Tailwind in `web/tailwind.config.ts`. Light/dark with a persisted toggle; i18n in en/es/pt-BR.

## Known limitations
- Payment is simulated (the saga's confirm step always succeeds unless told otherwise).
- Single MongoDB instance shared by services (logical DB-per-service, not physical isolation).
- No auth/users yet — the engine focuses on the inventory/ordering path.
- Inter-service calls are synchronous HTTP; no message bus.
