<div align="center">

# archive-commerce

A multi-store ecommerce builder (SaaS, à la Tienda Nube / Shopify) built around one hard problem:
never oversell inventory under concurrent load — now isolated per store. Merchants sign up, each
builds their own storefront and loads their own products. Microservices in Go, a SvelteKit
marketing site + per-store storefront + merchant dashboard, MongoDB and Redis.

<a href="#quickstart">Quickstart</a> · <a href="#the-concurrency-demo">Concurrency demo</a> · <a href="ARCHITECTURE.md">Architecture</a>

</div>

<!-- Add a capture of the running storefront at docs/screenshots/storefront.png and reference it here. -->

## The problem

When a thousand shoppers race for the last ten units, a naive store sells twelve. archive-commerce
holds the line: stock moves through atomic, guarded operations in MongoDB, so the available count
never goes negative and exactly the available number of orders succeed — the rest get a clean
"out of stock". Carts get a real, time-boxed hold on stock; abandoned carts return their units
automatically.

## Stack

| Layer | Choice | Why |
|-------|--------|-----|
| Backend | Go 1.22, one module → 4 services | Goroutines and the race detector make concurrency claims testable; a single module shares a `platform` package without four repos |
| Datastore | MongoDB | Document-level atomic updates (`findOneAndUpdate` with a guard) solve the oversell problem without app locks or distributed transactions |
| Cache / limiter | Redis | Distributed sliding-window rate limiting at the gateway and cache-aside reads for the catalog |
| Storefront | SvelteKit 5 + TypeScript + Tailwind | SSR reads over the internal network; small bundle; runes |
| Orchestration | Docker Compose | One command brings up the full stack |

## Features

- **No oversell under load** — atomic per-SKU reservations; a load test fires hundreds of concurrent
  buyers at a scarce SKU and asserts zero oversell.
- **Time-boxed reservations with compensation** — a background janitor returns stock from expired
  holds (a TTL index alone would only delete the record and leak the stock).
- **Checkout saga** — orders reserve, persist, then confirm; any failure triggers a compensating
  release. No external orchestrator.
- **Server-authoritative pricing** — totals are computed from the catalog, never trusted from the client.
- **API gateway** — single entry point with reverse proxying, CORS and Redis sliding-window rate limiting.
- **Editorial storefront** — a monochrome "refined brutalism" design system, light/dark, in English,
  Spanish and Brazilian Portuguese.
- **Multi-store SaaS** — a marketing landing at `/`, merchant onboarding (`/signup` creates the account
  and the first store), per-store storefront at `/store/[slug]` (subdomain in production), and a merchant
  dashboard at `/app` (products, stock, branding). Every store is isolated by `tenantId`.

## Quickstart

Requires Docker. Go is not needed locally — everything builds in containers.

```bash
cp .env.example .env
docker compose up --build
```

- Marketing landing: http://localhost:3000
- Demo storefront: http://localhost:3000/store/system-archive
- Merchant dashboard: http://localhost:3000/app (sign in at `/login`, or create a store at `/signup`)
- API gateway: http://localhost:8080

```bash
# Health (gateway aggregates each service)
curl http://localhost:8080/health

# Browse the seeded catalog
curl http://localhost:8080/api/catalog/products

# Place an order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{"items":[{"sku":"UR-OS","qty":1}]}'
```

The catalog and stock seed themselves on first boot.

## The concurrency demo

```bash
docker compose run --rm loadtest
```

Hundreds of buyers compete for a SKU stocked at 10 units. The load test resets stock to seed
defaults first, so it produces the same result on every run. Sample:

```
Target SKU AX1-85 — starting available: 10
Firing 500 buyers (max 150 in flight)...

Results
-------
  confirmed orders : 10
  out of stock     : 490
  other failures   : 0
  stock available  : 10 -> 0

OK: sold exactly 10 of 10 units, zero oversell.
```

## Tests

```bash
# Go services, with the race detector
docker compose run --rm backend-test

# Storefront unit tests
cd web && npm install && npm test
```

The inventory suite covers the central guarantee (no oversell across hundreds of goroutines),
confirm/release transitions, multi-SKU compensation and janitor expiry.

## Project layout

```
archive-commerce/
├── services/                 Go module, one binary per service
│   ├── cmd/{gateway,catalog,inventory,orders,accounts,loadtest}/
│   └── internal/
│       ├── platform/         config, mongo, redis, http + tenant + authx (JWT/bcrypt)
│       ├── accounts/         merchant auth + per-store (tenant) registry
│       ├── catalog/          products + Redis cache-aside (per tenant)
│       ├── inventory/        atomic stock, reservations, janitor (concurrency core, per tenant)
│       └── orders/           checkout saga
├── web/                      SvelteKit: landing + per-store storefront + merchant dashboard
├── docker-compose.yml
└── ARCHITECTURE.md
```

## Deployment

The stack runs anywhere Docker does. For a VPS, put Nginx in front of the gateway (`:8080`) and the
storefront (`:3000`) with Let's Encrypt for TLS; keep MongoDB and Redis on the internal Docker
network, not exposed to the public.

## Known limitations

- Payment is simulated; the saga's confirm step succeeds unless the upstream call fails.
- Inter-service calls are synchronous HTTP — an event-driven outbox is the natural next step.
- One MongoDB instance, one logical database per service (no physical isolation).
- Stores share collections, isolated logically by `tenantId` (Shopify/Tienda Nube model), not physically.
- No billing/plans yet, and the platform super-admin (`/admin`) is a stub for a later stage.
