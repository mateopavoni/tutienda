<div align="center">

# TuTienda

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

### CI/CD (one push, full deploy)

A push to `main` runs `.github/workflows/deploy.yml`, which after `go test -race` deploys **everything
in one step**:

1. **Backend** — over SSH the VPS runs `git pull --ff-only && docker compose up -d --build`, rebuilding
   the compose services (`accounts`, `catalog`, `inventory`, `orders` + the compose `gateway`/`web`).
   Mongo/Redis/MinIO volumes are **not** touched, so data persists and **no seed runs**.
2. **Public** — `gateway` (`tutienda-api`) and `storefront` (`tutienda-web`) ship to Dokku (TLS).

The public site is served by the Dokku apps; they reach the backend by Docker DNS (`accounts:8084`, …)
because the Dokku containers share the compose network (`archive-commerce_default`).

One-time VPS + GitHub setup (so the backend step works):

```bash
# On the VPS, make ~/tutienda a git checkout that can pull passwordless via a read-only deploy key:
cd ~/tutienda
ssh-keygen -t ed25519 -f ~/.ssh/gha -N "" -C "github-actions"   # one keypair, used both ways
cat ~/.ssh/gha.pub >> ~/.ssh/authorized_keys                    # lets GitHub Actions SSH in as ubuntu
printf 'Host github.com\n  IdentityFile ~/.ssh/gha\n  IdentitiesOnly yes\n' >> ~/.ssh/config
git remote set-url origin git@github.com:mateopavoni/tutienda.git
git fetch origin && git checkout -B main origin/main            # track origin/main for --ff-only pulls
```

Then in GitHub: add `~/.ssh/gha.pub` as a **Deploy key** (read-only) on the repo, and add the **private**
key `~/.ssh/gha` as the Actions secret **`VPS_SSH_PRIVATE_KEY`**. (`DOKKU_SSH_PRIVATE_KEY` stays as-is.)

> ⚠️ **Seeds and data.** The seed (demo + example stores) only runs on an **empty** database — it is a
> no-op once any merchant exists. The CI deploy never wipes data, so re-seeding requires clearing the
> Mongo volume by hand. Never run `docker compose down -v` once real customer/store data exists: `-v`
> deletes the Mongo/Redis/MinIO volumes. To re-seed selectively, drop only the seed documents.

## Known limitations

- Payment is simulated; the saga's confirm step succeeds unless the upstream call fails.
- Inter-service calls are synchronous HTTP — an event-driven outbox is the natural next step.
- One MongoDB instance, one logical database per service (no physical isolation).
- Stores share collections, isolated logically by `tenantId` (Shopify/Tienda Nube model), not physically.
- No billing/plans yet, and the platform super-admin (`/admin`) is a stub for a later stage.

## License

© 2026 Mateo Pavoni. All rights reserved. This is proprietary, closed-source software made
available for viewing/evaluation purposes only. Copying, redistribution, or reuse of any part of
this codebase, design, or branding without prior written permission is strictly prohibited. See
[LICENSE](./LICENSE).
