# Problemas conocidos

- **FACT** El pago es simulado (`approve`); no hay pasarela real.
- **FACT** Llamadas entre servicios por HTTP síncrono; sin outbox ni cola.
- **FACT** Una sola instancia de MongoDB y tenants en colecciones compartidas.
- **FACT** `gateway` y `orders` no tienen tests unitarios; la cobertura de la seguridad del gateway se verificó a mano contra el stack.
- **FACT** El rate limit por defecto (100 req/min por IP) corta scripts de prueba rápido; subir `RATE_LIMIT_MAX` en `.env` para pruebas.
- **FACT** La guard de prod solo rechaza el `JWT_SECRET` por defecto del código (`dev-insecure-secret-change-me`); el valor de ejemplo de `.env.example` (`change-me-to-a-long-random-secret`) sí pasa la guard. Hay que rotarlo a mano.
- **FACT** `web/static/{robots.txt,llms.txt}`, el sitemap y el JSON-LD de `app.html` siguen apuntando al dominio de la demo dada de baja.
- **FACT** Hay credenciales demo públicas sembradas (`demo@system-archive.store`, `admin@tutienda.store`) cuando `SEED_DEMO_DATA` no es false.
- **FACT** GitHub puede conservar secrets `DOKKU_*`, `VPS_SSH_PRIVATE_KEY` del deploy anterior; no se borraron (no autorizado).
