# Roadmap (si se retoma)

1. Tests de gateway y orders (hoy cubiertos solo por verificación manual).
2. Outbox + cola para la saga de checkout en lugar de HTTP síncrono.
3. Pasarela de pago real (Mercado Pago / Stripe) en lugar de `approve`.
4. Endurecer la guard de prod: rechazar también placeholders conocidos de `.env.example`.
5. Playwright e2e en CI contra el stack de compose.
6. CDN/edge cache delante del storefront (el TTFB era el mayor lastre de performance).
