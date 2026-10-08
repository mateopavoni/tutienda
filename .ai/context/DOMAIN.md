# Dominio

- **FACT** Merchant (cuenta) → N Stores (`slug`, `displayName`, `settings`, plan). Store = tenant; todo documento lleva `tenantId`.
- **FACT** Customer: cuenta de comprador por tienda (`/api/customers/{signup,login,me}`); sus pedidos solo los ve él mismo.
- **FACT** Product con Variants (cada variant tiene `sku` y `label`; el precio vive en el producto, `priceCents`). Stock por SKU: `{available, reserved}`.
- **FACT** Order: items + dirección de envío (obligatoria) + `approve` (simula el resultado del pago; no hay pasarela real).
- **FACT** Tienda demo `system-archive` (tenant fijo `000000000000000000000001`) más 3 tiendas de ejemplo, sembradas si la base está vacía y `SEED_DEMO_DATA` no es false.
- **FACT** Planes limitan recursos por tienda (`internal/platform/plan`).
