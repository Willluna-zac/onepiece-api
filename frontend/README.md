# One Piece — Frontend

Cliente React + TypeScript para la One Piece API. Construye el mapa interactivo del Grand Line con búsquedas Dijkstra en 4 modos.

## Stack

- **React 18** + **TypeScript** (strict)
- **Vite 5** — dev server con HMR + proxy `/api/*` al backend Go en `:8080`
- **TanStack Query (React Query)** — fetching y cache
- **React Router DOM** — navegación
- **Tailwind CSS** — estilos
- **Vitest** + **@testing-library/react** — tests

## Quick start

```bash
npm install
npm run dev          # http://localhost:5173 (proxy a backend en :8080)
npm test -- --run    # 40 tests
npm run build        # producción → dist/
```

Asegurate de tener el backend corriendo (`go run .` desde la raíz). Si querés apuntar a otro host, definí `VITE_API_URL` en `.env`.

## Estructura

```text
src/
  api/client.ts        # tipos + funciones fetch agrupadas por recurso
  hooks/               # wrappers de useQuery + useLastRouteSearch + useCompareModes
  components/
    RouteMap.tsx       # mapa SVG con tooltips, animación, anillos de peligro
    Navbar.tsx, Loader.tsx, CharacterCard.tsx
  lib/format.ts        # formatHours()
  pages/
    RoutesPage.tsx     # 4 modos + 4 métricas + comparativa + animación
    CharactersPage.tsx, IslandsPage.tsx, DevilFruitsPage.tsx, CharacterDetailPage.tsx
```

## Vista de Rutas (v2)

[`/routes`](http://localhost:5173/routes) ofrece:

- **Toggle 4 modos**: `fastest` / `quickest` / `safest` / `riskiest`. Navegable con ←/→ del teclado (roving tabindex).
- **4 métricas globales** siempre visibles (Distancia / Tiempo / Peor ☠ / Mejor ☠), con la del modo activo destacada en oro.
- **Mapa SVG** con path resaltado, anillos de peligro derivado por isla, tooltips enriquecidos en aristas (distancia / tiempo / peligro) y en islas (región / Log Pose / descripción).
- **Animación MVP** de la nave (círculo dorado avanzando por el path en 5s, `<animateMotion>` SVG).
- **Comparativa de los 4 modos** (4 queries en paralelo con `useQueries`, cache hit gratis del modo activo). Avisa cuando los 4 modos convergen al mismo path.
- **Persistencia** de la última búsqueda `{from, to, mode}` en `localStorage` (modo "solo restaurar": rellena los selects pero no auto-busca).
- **Accesibilidad**: `aria-live="polite"` en el contenedor de resultado, `aria-pressed` en los botones de modo, foco visible.

## Vista Análisis del grafo (v3)

[`/stats`](http://localhost:5173/stats) consume `GET /api/routes/stats` y muestra:

- **Tarjetas de totales**: islas, rutas, componentes conexos (✅ saludable / ⚠️ con aviso del componente mayor) y bidireccionales.
- **Promedios** de distancia, tiempo y danger.
- **Histograma de Danger** (1–5) renderizado con divs+colores semaforizados (sin librerías externas).
- **Notas finales** con porcentaje de islas con Log Pose y diagnóstico de conectividad.

Hook: `useGraphStats()` (`useQuery`, `staleTime: 60_000`) — alineado con el `Cache-Control: public, max-age=60` del backend.

## Tests

```bash
npm test -- --run                    # corre todo
npm test -- --run RoutesPage         # filtro por archivo
npm test -- --run --coverage         # coverage
```

Cobertura actual: **43 tests** distribuidos en `lib/`, `api/`, `hooks/`, `components/`, `pages/` (incluye `StatsPage`).

