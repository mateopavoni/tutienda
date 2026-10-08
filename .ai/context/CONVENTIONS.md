# Convenciones

- **FACT** Todo local: `cp .env.example .env && docker compose up --build` (web `127.0.0.1:3000`, gateway `127.0.0.1:8080`, MinIO `127.0.0.1:9000/9001`; Mongo y Redis sin publicar).
- **FACT** Sin Go/Node asumidos en el host: tests con contenedores (`golang:1.23`, `node:22`) o los servicios de compose `backend-test` y `loadtest`.
- **FACT** Tests: `services` → `go vet ./... && go test -race ./...` (inventory prende tests de integración si hay `MONGO_URI`); `web` → `npm run check && npm test && npm run build`.
- **FACT** Demo de concurrencia: `docker compose run --rm loadtest` (500 compradores, 10 unidades → 10 ventas exactas).
- **FACT** CI: `.github/workflows/ci.yml` (api + web) y `guard-coauthor.yml`. No hay jobs de deploy.
- **FACT** Commits como `Mateo Pavoni <mateopavoni905@gmail.com>`, sin trailers de Claude (hook `.githooks/commit-msg`; activar con `git config core.hooksPath .githooks`).
- **FACT** Mensajes de commit estilo convencional (`fix:`, `feat:`, `docs:`, `chore:`).
