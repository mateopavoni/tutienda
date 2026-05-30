# Design tokens — archive-commerce ("SYSTEM ARCHIVE")

Fuente de verdad: export de Stitch AI (`docs/design/stitch/`). Lenguaje visual:
**Refined Brutalism** — monocromático, plano, sin sombras, radio 0, separación por borde 1px.
El producto es el único color orgánico; la UI es una grilla técnica de "spec-sheet".

> Nota sobre tipografía: el dúo **Inter + JetBrains Mono** es una decisión deliberada del design
> system (titulares Swiss de escala extrema + metadata monoespaciada estilo ficha técnica), no un
> default. Está documentada acá para que sea defendible y no se confunda con la fuente genérica.

## Principios estructurales
| Token            | Valor   | Regla |
|------------------|---------|-------|
| radius           | `0`     | Hard-edge: ningún borde redondeado en todo el sistema |
| border-width     | `1px`   | "The 1px rule": toda separación es un borde onyx de 1px |
| elevation        | ninguna | Plano total. La profundidad se comunica por color blocking, no por sombras |
| grid             | 12 col  | Colapsa a 2 col en mobile; gutter de 1px |
| spacing-unit     | `4px`   | Escala 4/8/12/16/24/32/48/64 |
| container-max    | `1440px`| Margen 48px desktop / 16px mobile |

## Colores — pares claro / oscuro (semánticos)
El sistema es monocromático + 1 acento (Electric Ultramarine), reservado a conversión y feedback crítico.

| Token semántico   | Light    | Dark     | Uso |
|-------------------|----------|----------|-----|
| `bg`              | `#faf9f9`| `#111111`| Fondo principal (Alabaster / Onyx) |
| `surface`         | `#ffffff`| `#1b1b1b`| Cards, paneles, inputs |
| `surface-variant` | `#e3e2e2`| `#262626`| Fondos de imagen, hover sutil |
| `border`          | `#1b1c1c`| `#c4c7c7`| Bordes 1px (regla central) |
| `text`            | `#1b1c1c`| `#f2f0f0`| Texto principal (Onyx / Alabaster) |
| `text-muted`      | `#444748`| `#a8abab`| Metadata, secundario |
| `outline`         | `#747878`| `#5f5e5e`| Bordes inactivos, disabled |
| `primary`         | `#000000`| `#ffffff`| Superficies de alto impacto, CTAs sólidos |
| `on-primary`      | `#ffffff`| `#111111`| Texto sobre primary |
| `accent`          | `#1b03ea`| `#7073ff`| Electric Ultramarine — solo conversión (checkout) y feedback |
| `on-accent`       | `#ffffff`| `#ffffff`| Texto sobre accent |
| `error`           | `#ba1a1a`| `#ffb4ab`| Errores, out-of-stock |

Modo inicial: `prefers-color-scheme`; toggle manual persistido en `localStorage`. La pantalla
`cart_drawer_view` del export está en dark — el drawer usa siempre fondo de alto contraste.

## Tipografía
| Token            | Familia        | Size / line / weight / tracking |
|------------------|----------------|---------------------------------|
| `display-xl`     | Inter          | 120 / 100 / 800 / -0.04em |
| `headline-lg`    | Inter          | 64 / 64 / 700 / -0.02em |
| `headline-md`    | Inter          | 32 / 36 / 600 / -0.01em |
| `body-lg`        | Inter          | 18 / 28 / 400 |
| `body-md`        | Inter          | 16 / 24 / 400 |
| `metadata-sm`    | JetBrains Mono | 12 / 16 / 500 / 0.05em — SKU, precios, dimensiones |
| `label-caps`     | Inter          | 11 / 12 / 700 / 0.1em — nav, categorías (uppercase) |

## Componentes (patrones del export)
- **Botón primario:** rectángulo sólido onyx o ultramarine, texto Alabaster en mayúsculas; hover = inversión de color.
- **Botón secundario:** Alabaster con borde 1px onyx.
- **Input:** wireframe — borde inferior 1px inactivo → caja 1px completa al focus; label en `metadata-sm`.
- **Product card:** perímetro 1px, imagen sin padding al ancho de la card, metadata separada por línea 1px horizontal. Imagen en `grayscale` → color al hover.
- **Selector de variante:** fila de cajas 1px; seleccionado = relleno onyx sólido; out-of-stock = tachado en diagonal.
- **Cart drawer:** panel de alto contraste que entra desde la derecha; botón de checkout en ultramarine.
- **Iconos:** stroke-based (Material Symbols Outlined en el export; en la app usamos `lucide-svelte`, stroke 1.5–2px, joins rectos).

## Mapeo al stack (SvelteKit + Tailwind)
- `web/tailwind.config.ts` → `theme.extend` con estos tokens (colors, fontFamily, fontSize, spacing, borderRadius=0).
- `darkMode: 'class'` con clase `.dark` en `<html>`.
- Variables semánticas como CSS custom properties en `web/src/app.css` para ambos modos.
