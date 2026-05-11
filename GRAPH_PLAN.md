# Plan de Implementacion: Grafo de Rutas Maritimas + Dijkstra

## Decisiones previas

### IDs de islas: slug vs numerico
El plan original usa IDs numericos (1, 2, 3...) pero el repo actual usa slugs:

| # Plan | Slug actual           | Isla                  |
|--------|-----------------------|-----------------------|
| 1      | windmill-village      | Windmill Village      |
| 2      | shells-town           | Shells Town           |
| 3      | orange-town           | Orange Town           |
| 4      | syrup-village         | Syrup Village         |
| 5      | baratie               | Baratie               |
| —      | arlong-park           | Arlong Park           |
| 6      | loguetown             | Loguetown             |
| 7      | reverse-mountain      | Reverse Mountain      |
| 8      | whiskey-peak          | Whiskey Peak          |
| 9      | little-garden         | Little Garden         |
| —      | drum-island           | Drum Island           |
| 10     | alabasta              | Alabasta              |
| 11     | jaya                  | Jaya                  |
| 12     | skypiea               | Skypiea               |
| 13     | long-ring-long-land   | Long Ring Long Land   |
| 14     | water-7               | Water 7               |
| 15     | enies-lobby           | Enies Lobby           |
| 16     | thriller-bark         | Thriller Bark         |
| 17     | sabaody-archipelago   | Sabaody Archipelago   |
| 18     | amazon-lily           | Amazon Lily           |
| 19     | impel-down            | Impel Down            |
| 20     | marineford            | Marineford            |
| 21     | fishman-island        | Fish-Man Island       |
| 22     | punk-hazard           | Punk Hazard           |
| 23     | dressrosa             | Dressrosa             |
| 24     | zou                   | Zou                   |
| 25     | whole-cake-island     | Whole Cake Island     |
| 26     | wano                  | Wano Country          |
| 27     | elbaf                 | Elbaf                 |
| 28     | laugh-tale            | Laugh Tale            |
| 29     | —                     | Kano Country (NO EXISTE) |
| 30     | —                     | Banaro Island (NO EXISTE) |

**Decision:** Agregar Kano Country y Banaro Island al island_repository.go, y agregar rutas
para Arlong Park y Drum Island para que todas las islas sean alcanzables en el grafo.
Esto nos da 32 islas como nodos del grafo.

Rutas adicionales sugeridas para las islas huerfanas:
- Baratie <-> Arlong Park (distancia 224, peligro 2, bidireccional) — Nami va y vuelve
- Whiskey Peak <-> Drum Island (distancia 400, peligro 3, bidireccional) — ruta del Log Pose
- Drum Island <-> Alabasta (distancia 300, peligro 2, bidireccional) — siguiente parada canonica

---

## Checklist de implementacion

### Paso 1: pkg/graph/graph.go — Estructura generica del grafo
- [x] Crear directorio `pkg/graph/`
- [x] Implementar `Edge` struct (To string, Weight float64)
- [x] Implementar `Graph` struct (adjacency map[string][]Edge, nodes map[string]bool)
- [x] Implementar `New() *Graph`
- [x] Implementar `AddNode(id string)`
- [x] Implementar `AddEdge(from, to string, weight float64)`
- [x] Implementar `AddBidirectionalEdge(from, to string, weight float64)`
- [x] Implementar `Neighbors(id string) []Edge`
- [x] Implementar `Nodes() []string`
- [x] Implementar `HasNode(id string) bool`
- [x] Implementar `EdgeCount() int`

**Archivo:** `pkg/graph/graph.go`
**Dependencias:** Ninguna (Go puro)

---

### Paso 2: pkg/graph/dijkstra.go — Algoritmo de Dijkstra
- [x] Implementar `PathResult` struct (Path []string, TotalCost float64, Found bool)
- [x] Implementar `priorityQueue` (min-heap usando container/heap)
  - [x] `pqItem` struct (node, cost, index)
  - [x] Len, Less, Swap, Push, Pop
- [x] Implementar `Dijkstra(source, target string) PathResult`
  - [x] Manejar nodos inexistentes (return Found: false)
  - [x] Manejar source == target (return Found: true, cost 0)
  - [x] Min-heap + relajacion de aristas
  - [x] Reconstruccion del camino con `reconstructPath`
- [x] Implementar `DijkstraAll(source string) map[string]float64`
  - [x] Distancias desde source a TODOS los nodos
  - [x] Nodos inalcanzables quedan en math.Inf(1)
- [x] Implementar `reconstructPath(prev, source, target) []string`

**Archivo:** `pkg/graph/dijkstra.go`
**Dependencias:** container/heap, math

---

### Paso 3: pkg/graph/graph_test.go — Tests unitarios
- [x] TestDijkstra_SimpleGraph — A->B->C mas barato que A->C directo
- [x] TestDijkstra_DirectedGraph — Reverse Mountain solo ida, no hay retorno
- [x] TestDijkstra_Unreachable — nodos desconectados
- [x] TestDijkstra_SameNode — costo 0 a si mismo
- [x] TestDijkstra_NonexistentNode — nodo que no existe
- [x] TestDijkstraAll — distancias a todos, incluido nodo aislado (Inf)
- [x] TestDijkstraAll_NonexistentSource — source inexistente retorna todo Inf
- [x] TestGraph_EdgeCount — bidireccional cuenta como 2
- [x] TestGraph_AddNodeIdempotent — agregar nodo repetido no duplica
- [x] TestGraph_HasNode — existe vs no existe
- [x] TestGraph_Neighbors — aristas salientes, nodo sin vecinos, nodo inexistente
- [x] Ejecutar `go test ./pkg/graph/...` — 11/11 PASS

**Archivo:** `pkg/graph/graph_test.go`
**Dependencias:** Paso 1, Paso 2

---

### Paso 4: domain/route.go — Modelo de dominio
- [x] `Route` struct con campos: ID, FromIsland, ToIsland, Distance, Danger, Weight, Bidirectional, Notes
  - [x] Tags JSON (camelCase)
- [x] `ShortestPathResponse` struct: From, To, TotalCost, Hops, Path []PathStep, Found
- [x] `PathStep` struct: IslandID, IslandName, CostSoFar
- [x] `ReachableResponse` struct: From, MaxCost, Islands []ReachableIsland
- [x] `ReachableIsland` struct: IslandID, IslandName, Cost
- [x] `DangerMultiplier(danger int) float64` — retorna multiplicador (1->1.0, 2->1.3, 3->1.6, 4->2.0, 5->3.0)

**Archivo:** `domain/route.go`
**Dependencias:** Ninguna

---

### Paso 5: domain/ports.go — Interfaces
- [x] Agregar `RouteRepository` interface:
  - [x] `GetAll(ctx context.Context) ([]Route, error)`
  - [x] `GetByIsland(ctx context.Context, islandID string) ([]Route, error)`
- [x] Agregar `RouteUseCase` interface:
  - [x] `GetAllRoutes(ctx context.Context) ([]Route, error)`
  - [x] `GetRoutesFromIsland(ctx context.Context, islandID string) ([]Route, error)`
  - [x] `FindShortestPath(ctx context.Context, fromID, toID string) (*ShortestPathResponse, error)`
  - [x] `FindReachableIslands(ctx context.Context, fromID string, maxCost float64) (*ReachableResponse, error)`

**Archivo:** `domain/ports.go` (agregar al existente)
**Dependencias:** Paso 4 (necesita los tipos Route, ShortestPathResponse, etc.)

---

### Paso 6: repository/route_repository.go — Datos de rutas maritimas
- [x] Agregar 2 islas nuevas al island_repository.go:
  - [x] Kano Country (slug: kano-country, region: North Blue, X: 800, Y: 3200)
  - [x] Banaro Island (slug: banaro-island, region: Grand Line, X: 3400, Y: 1500)
- [x] Crear `RouteRepository` struct con `routes []domain.Route`
- [x] Crear `NewRouteRepository() *RouteRepository` con seed de todas las rutas
- [x] Implementar `GetAll` — retorna todas las rutas
- [x] Implementar `GetByIsland` — filtra rutas donde FromIsland o ToIsland == islandID
- [x] Calcular Weight = Distance * DangerMultiplier(Danger) para cada ruta
- [x] Hardcodear las 37 rutas del plan (con slugs) + 3 rutas extra:
  - [ ] East Blue (rutas 1-6): windmill-village <-> shells-town <-> orange-town <-> syrup-village <-> baratie <-> loguetown, windmill-village <-> kano-country, baratie <-> arlong-park
  - [ ] Entrada Grand Line (rutas 7-8): loguetown -> reverse-mountain -> whiskey-peak (solo ida)
  - [ ] Paradise (rutas 9-17): whiskey-peak <-> little-garden <-> alabasta <-> jaya, jaya -> skypiea (solo ida), jaya <-> long-ring-long-land <-> water-7, water-7 <-> enies-lobby, water-7 <-> thriller-bark <-> sabaody-archipelago
  - [ ] Rutas alternativas Paradise (rutas 18-19): alabasta <-> long-ring-long-land, whiskey-peak <-> alabasta
  - [ ] Red Line y submarina (rutas 20-23): sabaody -> fishman-island, sabaody <-> marineford, marineford <-> impel-down, impel-down <-> amazon-lily
  - [ ] New World (rutas 24-32): fishman-island <-> punk-hazard <-> dressrosa <-> zou, zou <-> whole-cake-island, zou <-> wano, whole-cake-island <-> wano, wano <-> elbaf -> laugh-tale (solo ida), wano -> laugh-tale (solo ida)
  - [ ] Conexiones especiales (rutas 33-37+): skypiea -> jaya (solo bajada), banaro-island <-> alabasta, amazon-lily <-> fishman-island, kano-country <-> reverse-mountain, enies-lobby <-> sabaody-archipelago
  - [ ] Rutas extra: whiskey-peak <-> drum-island, drum-island <-> alabasta

**Archivo:** `repository/route_repository.go` (nuevo), `repository/island_repository.go` (agregar 2 islas)
**Dependencias:** Paso 4

---

### Paso 7: usecase/route_usecase.go — Logica de negocio
- [x] Crear `routeUseCase` struct con routeRepo y islandRepo
- [x] Crear `NewRouteUseCase(rr domain.RouteRepository, ir IslandRepository) domain.RouteUseCase`
- [x] Implementar `GetAllRoutes` — delega al repo
- [x] Implementar `GetRoutesFromIsland` — delega al repo
- [x] Implementar `FindShortestPath(ctx, fromID, toID)`:
  - [x] Obtener todas las rutas del repo
  - [x] Construir grafo con `graph.New()`
  - [x] Iterar rutas: AddEdge o AddBidirectionalEdge segun r.Bidirectional
  - [x] Ejecutar `g.Dijkstra(fromID, toID)`
  - [x] Si no encontro camino: retornar Found: false
  - [x] Enriquecer path con nombres de isla (obtener islas del islandRepo)
  - [x] Calcular costSoFar acumulado para cada paso
  - [x] Retornar ShortestPathResponse completo
- [x] Implementar `FindReachableIslands(ctx, fromID, maxCost)`:
  - [x] Obtener todas las rutas, construir grafo
  - [x] Ejecutar `g.DijkstraAll(fromID)`
  - [x] Filtrar nodos con distancia <= maxCost (excluir Inf)
  - [x] Enriquecer con nombres de isla
  - [x] Ordenar por costo ascendente
  - [x] Retornar ReachableResponse

**Archivo:** `usecase/route_usecase.go`
**Dependencias:** Paso 1-2 (graph), Paso 4-5 (domain), Paso 6 (repo)

---

### Paso 8: controller/route_controller.go — HTTP handlers
- [x] Crear `routeUsecaseInterface` (interfaz local como hace island_controller.go)
- [x] Crear `RouteController` struct con usecase
- [x] Crear `NewRouteController(uc routeUsecaseInterface) *RouteController`
- [x] Implementar `GetAllRoutes(w, r)` — GET /api/routes
- [x] Implementar `GetRoutesFromIsland(w, r)` — GET /api/routes/from/{islandID}
  - [x] Extraer islandID del path con strings.TrimPrefix
- [x] Implementar `GetShortestPath(w, r)` — GET /api/routes/shortest?from=&to=
  - [x] Validar query params "from" y "to"
  - [x] Llamar usecase.FindShortestPath
- [x] Implementar `GetReachableIslands(w, r)` — GET /api/routes/reachable?from=&maxCost=
  - [x] Validar query params "from" y "maxCost"
  - [x] Parsear maxCost como float64
  - [x] Llamar usecase.FindReachableIslands
- [x] Usar helpers existentes: sendError, sendJSON

**Archivo:** `controller/route_controller.go`
**Dependencias:** Paso 7

---

### Paso 9: router/router.go + main.go — Wiring
- [x] Modificar `SetupRoutes` para recibir `*controller.RouteController`
- [x] Registrar rutas (orden: especificas antes de wildcards):
  - [x] `GET /api/routes` — GetAllRoutes
  - [x] `GET /api/routes/shortest` — GetShortestPath
  - [x] `GET /api/routes/reachable` — GetReachableIslands
  - [x] `GET /api/routes/from/` — GetRoutesFromIsland (wildcard)
- [x] En main.go agregar DI:
  - [x] `routeRepo := repository.NewRouteRepository()`
  - [x] `routeUsecase := usecase.NewRouteUseCase(routeRepo, islandRepo)`
  - [x] `routeController := controller.NewRouteController(routeUsecase)`
  - [x] Pasar routeController a SetupRoutes
- [x] Agregar prints de los nuevos endpoints en main.go

**Archivo:** `router/router.go`, `main.go`
**Dependencias:** Paso 8

---

### Paso 10: Verificacion final
- [x] `go build ./...` — compila sin errores
- [x] `go test ./pkg/graph/...` — todos los tests pasan
- [x] `go vet ./...` — sin warnings
- [x] Levantar servidor y probar manualmente:
  - [x] `curl localhost:8080/api/routes` — lista todas las rutas
  - [x] `curl localhost:8080/api/routes/from/loguetown` — rutas desde Loguetown
  - [x] `curl "localhost:8080/api/routes/shortest?from=windmill-village&to=wano"` — ruta mas corta
  - [x] `curl "localhost:8080/api/routes/reachable?from=loguetown&maxCost=1500"` — islas alcanzables
- [x] Verificar que endpoints existentes siguen funcionando (no regresion) — 32 islas OK

---

## Resumen del grafo final

- **Nodos:** 32 islas (30 originales + Kano Country + Banaro Island)
- **Aristas:** ~40 rutas (37 del plan + 3 extra para Arlong Park y Drum Island)
- **Aristas dirigidas:** ~70 (considerando bidireccionalidad)
- **Rutas solo ida:** Reverse Mountain -> Whiskey Peak, Jaya -> Skypiea, Skypiea -> Jaya (bajada), Wano -> Laugh Tale, Elbaf -> Laugh Tale
