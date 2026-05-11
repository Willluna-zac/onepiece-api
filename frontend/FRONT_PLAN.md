# Plan de Implementacion: Frontend — Rutas Maritimas y Grafo

## Stack y patrones del proyecto
- React 18 + TypeScript + Vite
- TanStack Query v5 (React Query)
- React Router DOM v6
- Tailwind CSS — sin librerías de grafos externas, SVG puro
- Patron: `api/client.ts` → `hooks/useApi.ts` → `pages/` + `components/`

## Decisiones de diseno
- **Sin librerías externas de grafos** (d3, cytoscape, reactflow) — SVG puro, consistente con lo actual
- **RouteMap** es stateless y reutilizable — el estado vive en las páginas
- **Colores por peligro:** danger 1→azul, 2→verde, 3→amarillo, 4→naranja, 5→rojo
- **Rutas solo-ida:** trazo discontinuo (`strokeDasharray`) para distinguirlas de bidireccionales
- **Tests:** Vitest + React Testing Library (el estándar del ecosistema Vite)

---

## Checklist de implementacion

### Paso 1: Setup de testing
- [x] Instalar dependencias: `vitest`, `@vitest/ui`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`, `jsdom`
- [x] Configurar `vitest.config.ts` con environment jsdom y globals
- [x] Agregar `"test": "vitest"` y `"test:ui": "vitest --ui"` a package.json scripts
- [x] Agregar `setupTests.ts` con import de `@testing-library/jest-dom`
- [x] Verificar con `npm test -- --run` (sin tests aun, debe pasar en 0 suites)

**Archivos:** `vitest.config.ts`, `src/setupTests.ts`, `package.json`

---

### Paso 2: api/client.ts — Tipos e interfaces de rutas
- [x] Agregar interface `Route`: id, fromIsland, toIsland, distance, danger, weight, bidirectional, notes
- [x] Agregar interface `PathStep`: islandId, islandName, costSoFar
- [x] Agregar interface `ShortestPathResponse`: from, to, totalCost, hops, path, found
- [x] Agregar interface `ReachableIsland`: islandId, islandName, cost
- [x] Agregar interface `ReachableResponse`: from, maxCost, islands
- [x] Agregar `routesApi` object:
  - [x] `getAll()` → GET /api/routes
  - [x] `fromIsland(id)` → GET /api/routes/from/{id}
  - [x] `shortest(from, to)` → GET /api/routes/shortest?from=&to=
  - [x] `reachable(from, maxCost)` → GET /api/routes/reachable?from=&maxCost=

**Archivo:** `src/api/client.ts`

---

### Paso 3: api/client.test.ts — Tests del cliente
- [x] `routesApi.getAll` llama a `/api/routes` y retorna array de Route
- [x] `routesApi.fromIsland` construye la URL correcta con el ID
- [x] `routesApi.shortest` construye query params `from` y `to` correctamente
- [x] `routesApi.reachable` construye query params `from` y `maxCost` correctamente
- [x] Ejecutar `npm test -- --run` — todos pasan

**Archivo:** `src/api/client.test.ts`
**Dependencias:** Paso 1, Paso 2

---

### Paso 4: hooks/useApi.ts — Hooks de rutas
- [x] Agregar `useRoutes()` — todas las rutas, staleTime: 300_000
- [x] Agregar `useRoutesFromIsland(islandID, enabled)` — rutas de una isla
- [x] Agregar `useShortestPath(from, to, enabled)` — enabled solo cuando ambos IDs definidos
- [x] Agregar `useReachableIslands(from, maxCost, enabled)` — enabled cuando from != "" y maxCost > 0

**Archivo:** `src/hooks/useApi.ts`
**Dependencias:** Paso 2

---

### Paso 5: hooks/useApi.test.ts — Tests de hooks
- [x] `useRoutes` llama a `routesApi.getAll` y expone data/isLoading/error
- [x] `useShortestPath` NO hace fetch cuando `enabled=false`
- [x] `useShortestPath` NO hace fetch cuando `from` está vacío
- [x] `useShortestPath` SI hace fetch cuando ambos IDs estan definidos
- [x] `useReachableIslands` NO hace fetch cuando maxCost = 0
- [x] `useReachableIslands` SI hace fetch con from y maxCost válidos
- [x] Ejecutar `npm test -- --run` — todos pasan

**Archivo:** `src/hooks/useApi.test.ts`
**Dependencias:** Paso 1, Paso 4

---

### Paso 6: components/RouteMap.tsx — Mapa SVG con rutas
- [x] Props: `islands: Island[]`, `routes: Route[]`, `highlightPath?: string[]`, `nearestIsland?: Island`, `onMapClick?: (x, y) => void`
- [x] Dibujar `<line>` SVG por cada ruta con color segun `danger` (1→azul, 2→verde, 3→amarillo, 4→naranja, 5→rojo)
- [x] Rutas bidireccionales: linea solida. Rutas solo-ida: `strokeDasharray="6 3"`
- [x] Si `highlightPath` definido: dibujar aristas del camino en dorado (#f5c842) con `strokeWidth=3`
- [x] Islas en el `highlightPath` con anillo dorado y radio mayor
- [x] Tooltip `<title>` en cada linea: "IslaA → IslaB | dist: X peligro: Y weight: Z"
- [x] Mantener click para isla mas cercana si `onMapClick` definido
- [x] Leyenda de colores de peligro en esquina inferior derecha

**Archivo:** `src/components/RouteMap.tsx`
**Dependencias:** Paso 2

---

### Paso 7: components/RouteMap.test.tsx — Tests del mapa
- [x] Renderiza sin rutas — solo islas como circulos SVG
- [x] Con rutas, renderiza el numero correcto de `<line>` elements
- [x] Con `highlightPath`, las aristas del path tienen stroke dorado
- [x] Ruta solo-ida tiene `strokeDasharray` en el SVG
- [x] Tooltip `<title>` contiene nombre de las islas
- [x] Ejecutar `npm test -- --run` — todos pasan

**Archivo:** `src/components/RouteMap.test.tsx`
**Dependencias:** Paso 1, Paso 6

---

### Paso 8: pages/RoutesPage.tsx — Pagina de navegacion
Dos secciones en la misma pagina:

#### Seccion A — Ruta mas corta (Dijkstra)
- [x] Dos `<select>` de isla poblados desde `useIslands()`: "Origen" y "Destino"
- [x] Boton "Calcular ruta" — disabled si falta alguno de los dos selects
- [x] Badge resultado: "✅ Ruta encontrada" o "❌ Sin ruta navegable"
- [x] Mostrar `totalCost` y `hops`
- [x] Timeline vertical: cada PathStep como tarjeta con islandName y costSoFar
- [x] `RouteMap` con `highlightPath` del resultado

#### Seccion B — Islas alcanzables
- [x] `<select>` de isla origen
- [x] Slider `maxCost` rango 100–10000, paso 100, valor default 1500
- [x] Mostrar valor actual del slider
- [x] Lista de islas alcanzables ordenadas por costo
- [x] Barra de progreso por isla: `(cost / maxCost) * 100%`
- [x] Badge de region en cada isla con color del sistema existente

**Archivo:** `src/pages/RoutesPage.tsx`
**Dependencias:** Paso 4, Paso 6

---

### Paso 9: pages/RoutesPage.test.tsx — Tests de la pagina
- [x] Renderiza ambas secciones ("Ruta mas corta" e "Islas alcanzables")
- [x] Boton "Calcular ruta" esta disabled cuando no hay selects elegidos
- [x] Slider de maxCost tiene valor inicial 1500
- [x] Con resultado mockeado de `useShortestPath`, muestra badge y paradas del path
- [x] Con `found: false`, muestra badge de "Sin ruta"
- [x] Ejecutar `npm test -- --run` — todos pasan

**Archivo:** `src/pages/RoutesPage.test.tsx`
**Dependencias:** Paso 1, Paso 8

---

### Paso 10: pages/IslandsPage.tsx — Actualizar mapa existente
- [x] Importar `RouteMap` y `useRoutes`
- [x] Reemplazar el componente `WorldMap` local por `<RouteMap>`
- [x] Pasar `routes` desde `useRoutes()` como prop
- [x] Mantener toda funcionalidad existente: click nearest, filtros de region, cards

**Archivo:** `src/pages/IslandsPage.tsx`
**Dependencias:** Paso 6

---

### Paso 11: components/Navbar.tsx + App.tsx — Wiring
- [x] `Navbar.tsx`: agregar `{ to: '/routes', label: '⚓ Rutas' }` al array `links`
- [x] `App.tsx`: importar `RoutesPage` y agregar `<Route path="/routes" element={<RoutesPage />} />`

**Archivos:** `src/components/Navbar.tsx`, `src/App.tsx`
**Dependencias:** Paso 8

---

### Paso 12: Verificacion final
- [x] `npm test -- --run` — 20/20 tests pasan
- [x] `npm run build` — compila sin errores TypeScript

---

## Resumen de archivos nuevos y modificados

| Archivo | Accion |
|---|---|
| `vitest.config.ts` | NUEVO |
| `src/setupTests.ts` | NUEVO |
| `src/api/client.ts` | MODIFICAR — agregar tipos y routesApi |
| `src/api/client.test.ts` | NUEVO |
| `src/hooks/useApi.ts` | MODIFICAR — agregar 4 hooks |
| `src/hooks/useApi.test.ts` | NUEVO |
| `src/components/RouteMap.tsx` | NUEVO |
| `src/components/RouteMap.test.tsx` | NUEVO |
| `src/pages/RoutesPage.tsx` | NUEVO |
| `src/pages/RoutesPage.test.tsx` | NUEVO |
| `src/pages/IslandsPage.tsx` | MODIFICAR — usar RouteMap |
| `src/components/Navbar.tsx` | MODIFICAR — agregar link |
| `src/App.tsx` | MODIFICAR — agregar ruta |
| `package.json` | MODIFICAR — agregar scripts test |
