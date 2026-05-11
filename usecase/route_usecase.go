package usecase

import (
	"context"
	"errors"
	"math"
	"sort"

	"onepiece-api/domain"
	"onepiece-api/pkg/graph"
)

// routeIslandRepository es el subconjunto de IslandRepository que necesita el use case de rutas.
type routeIslandRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Island, error)
	GetAll(ctx context.Context) ([]*domain.Island, error)
}

// RouteUsecase maneja la lógica de negocio de rutas marítimas y navegación.
type RouteUsecase struct {
	routeRepo  domain.RouteRepository
	islandRepo routeIslandRepository
}

// NewRouteUseCase crea una nueva instancia del use case de rutas.
func NewRouteUseCase(rr domain.RouteRepository, ir routeIslandRepository) *RouteUsecase {
	return &RouteUsecase{routeRepo: rr, islandRepo: ir}
}

// GetAllRoutes retorna todas las rutas marítimas registradas.
func (uc *RouteUsecase) GetAllRoutes(ctx context.Context) ([]domain.Route, error) {
	return uc.routeRepo.GetAll(ctx)
}

// GetRoutesFromIsland retorna todas las rutas donde la isla dada es origen o destino.
func (uc *RouteUsecase) GetRoutesFromIsland(ctx context.Context, islandID string) ([]domain.Route, error) {
	if islandID == "" {
		return nil, errors.New("el ID de isla no puede estar vacío")
	}
	return uc.routeRepo.GetByIsland(ctx, islandID)
}

// FindShortestPath calcula la ruta óptima entre dos islas según el modo solicitado.
//
//   - RouteModeFastest:  Dijkstra clásico (suma de Distance).
//   - RouteModeQuickest: Dijkstra clásico sobre tiempos (TravelHours + LogPose de
//     islas intermedias). LogPose del origen y destino NO cuentan.
//   - RouteModeSafest:   minimax sobre Danger (minimiza el peor tramo).
//   - RouteModeRiskiest: maximin sobre Danger (maximiza el mejor tramo).
//
// La respuesta SIEMPRE incluye las 4 métricas globales (TotalDistance, TotalTime,
// WorstDanger, BestDanger) calculadas sobre el camino encontrado, independientemente
// del modo. TotalCost (legacy) refleja la métrica del modo activo.
func (uc *RouteUsecase) FindShortestPath(ctx context.Context, fromID, toID string, mode domain.RouteMode) (*domain.ShortestPathResponse, error) {
	if fromID == "" || toID == "" {
		return nil, errors.New("los IDs de origen y destino son requeridos")
	}

	routes, err := uc.routeRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	islands, err := uc.islandRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	routesByEdge := indexRoutesByEdge(routes)
	islandsByID := indexIslandsByID(islands)

	g := buildGraphFromRoutes(routes, mode, islandsByID)

	var result graph.PathResult
	switch mode {
	case domain.RouteModeSafest:
		result = g.DijkstraBottleneck(fromID, toID, graph.BottleneckMin)
	case domain.RouteModeRiskiest:
		result = g.DijkstraBottleneck(fromID, toID, graph.BottleneckMax)
	default: // fastest, quickest
		result = g.Dijkstra(fromID, toID)
	}

	response := &domain.ShortestPathResponse{
		From:  fromID,
		To:    toID,
		Mode:  mode,
		Found: result.Found,
		Hops:  len(result.Path) - 1,
	}

	if !result.Found {
		return response, nil
	}

	steps, totals := enrichPath(result.Path, mode, routesByEdge, islandsByID)
	response.Path = steps
	response.TotalDistance = totals.distance
	response.TotalTime = totals.time
	response.WorstDanger = totals.worstDanger
	response.BestDanger = totals.bestDanger
	response.TotalCost = totalCostForMode(mode, totals)

	return response, nil
}

// FindReachableIslands retorna todas las islas alcanzables desde un origen con costo <= maxCost.
func (uc *RouteUsecase) FindReachableIslands(ctx context.Context, fromID string, maxCost float64) (*domain.ReachableResponse, error) {
	if fromID == "" {
		return nil, errors.New("el ID de isla origen es requerido")
	}
	if maxCost <= 0 {
		return nil, errors.New("maxCost debe ser mayor que 0")
	}

	routes, err := uc.routeRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	g := buildGraphFromRoutes(routes, domain.RouteModeFastest, nil)

	distances := g.DijkstraAll(fromID)

	var islands []domain.ReachableIsland
	for nodeID, cost := range distances {
		if nodeID == fromID || math.IsInf(cost, 1) || cost > maxCost {
			continue
		}
		island, err := uc.islandRepo.GetByID(ctx, nodeID)
		name := nodeID
		if err == nil {
			name = island.Name
		}
		islands = append(islands, domain.ReachableIsland{
			IslandID:   nodeID,
			IslandName: name,
			Cost:       cost,
		})
	}

	sort.Slice(islands, func(i, j int) bool {
		return islands[i].Cost < islands[j].Cost
	})

	return &domain.ReachableResponse{
		From:    fromID,
		MaxCost: maxCost,
		Islands: islands,
	}, nil
}

// ---------------------------------------------------------------------------
// helpers privados
// ---------------------------------------------------------------------------

// GetGraphStats calcula metricas estructurales del grafo a partir del snapshot
// actual del repositorio: totales, distribucion de Danger, promedios y numero
// de componentes conectados (BFS sobre el grafo no dirigido inducido por las
// rutas: las bidireccionales se recorren en ambos sentidos; las unidireccionales
// se cuentan solo en su sentido natural — un camino de regreso no garantizado
// puede dejar islas en componentes distintos, y eso es informacion util).
func (uc *RouteUsecase) GetGraphStats(ctx context.Context) (*domain.GraphStats, error) {
	routes, err := uc.routeRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	islands, err := uc.islandRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	stats := &domain.GraphStats{
		TotalIslands: len(islands),
		TotalRoutes:  len(routes),
	}

	for _, isl := range islands {
		if isl != nil && isl.LogPoseHours > 0 {
			stats.IslandsWithLogPose++
		}
	}

	var sumDist, sumTime, sumDanger float64
	for _, r := range routes {
		sumDist += r.Distance
		sumTime += r.TravelHours
		sumDanger += float64(r.Danger)
		if r.Bidirectional {
			stats.BidirectionalCount++
		}
		if r.Danger >= 1 && r.Danger <= 5 {
			stats.DangerHistogram[r.Danger-1]++
		}
	}
	if n := float64(len(routes)); n > 0 {
		stats.AvgDistance = sumDist / n
		stats.AvgTravelHours = sumTime / n
		stats.AvgDanger = sumDanger / n
	}

	// Componentes conectados (BFS). Solo se siguen aristas bidireccionales en
	// ambos sentidos; las unidireccionales solo en su direccion declarada.
	adj := make(map[string][]string, len(islands))
	for _, r := range routes {
		adj[r.FromIsland] = append(adj[r.FromIsland], r.ToIsland)
		if r.Bidirectional {
			adj[r.ToIsland] = append(adj[r.ToIsland], r.FromIsland)
		}
	}
	visited := make(map[string]bool, len(islands))
	for _, isl := range islands {
		if isl == nil || visited[isl.ID] {
			continue
		}
		size := 0
		queue := []string{isl.ID}
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			if visited[id] {
				continue
			}
			visited[id] = true
			size++
			queue = append(queue, adj[id]...)
		}
		stats.ConnectedComponents++
		if size > stats.LargestComponent {
			stats.LargestComponent = size
		}
	}

	return stats, nil
}

// pathTotals agrupa las 4 métricas globales del camino (siempre presentes
// en la respuesta, independientemente del modo optimizado).
type pathTotals struct {
	distance    float64
	time        float64
	worstDanger int
	bestDanger  int
}

// edgeKey produce una clave única para indexar una ruta dirigida.
func edgeKey(from, to string) string { return from + "→" + to }

// indexRoutesByEdge construye un map[edgeKey]Route con TODAS las direcciones
// efectivas: si la ruta es bidireccional se indexa también la inversa para que
// enrichPath pueda consultar por (from,to) sin importar el sentido del path.
func indexRoutesByEdge(routes []domain.Route) map[string]domain.Route {
	idx := make(map[string]domain.Route, len(routes)*2)
	for _, r := range routes {
		idx[edgeKey(r.FromIsland, r.ToIsland)] = r
		if r.Bidirectional {
			idx[edgeKey(r.ToIsland, r.FromIsland)] = r
		}
	}
	return idx
}

func indexIslandsByID(islands []*domain.Island) map[string]*domain.Island {
	idx := make(map[string]*domain.Island, len(islands))
	for _, isl := range islands {
		idx[isl.ID] = isl
	}
	return idx
}

// buildGraphFromRoutes construye el grafo eligiendo el peso de cada arista
// según el modo. Para `quickest` se aplica la transformación nodo→arista
// peso(u→v) = TravelHours(u→v) + LogPose(v); el LogPose del destino final se
// "descuenta" en enrichPath sumando solo LogPose de islas intermedias.
//
// Nota: el grafo se reconstruye en cada llamada. A la escala actual del
// dataset es despreciable; si crece habrá que cachearlo o invalidarlo por evento.
func buildGraphFromRoutes(routes []domain.Route, mode domain.RouteMode, islandsByID map[string]*domain.Island) *graph.Graph {
	g := graph.New()
	for _, r := range routes {
		wForward := edgeWeight(r, mode, islandsByID, r.ToIsland)
		if r.Bidirectional {
			wBackward := edgeWeight(r, mode, islandsByID, r.FromIsland)
			// AddBidirectionalEdge usa el mismo peso en ambos sentidos; cuando
			// los pesos difieren (caso quickest, depende de LogPose del destino)
			// agregamos las dos aristas dirigidas por separado.
			if wForward == wBackward {
				g.AddBidirectionalEdge(r.FromIsland, r.ToIsland, wForward)
			} else {
				g.AddEdge(r.FromIsland, r.ToIsland, wForward)
				g.AddEdge(r.ToIsland, r.FromIsland, wBackward)
			}
		} else {
			g.AddEdge(r.FromIsland, r.ToIsland, wForward)
		}
	}
	return g
}

// edgeWeight selecciona el atributo de la ruta que se usa como peso según el modo.
//
// Para `quickest`, el peso incluye el LogPoseHours de la isla DESTINO de la arista.
// Esto convierte el costo "de nodo" (esperar Log Pose) en costo "de arista entrante",
// permitiendo que Dijkstra encuentre el óptimo con una suma simple.
func edgeWeight(r domain.Route, mode domain.RouteMode, islandsByID map[string]*domain.Island, toIsland string) float64 {
	switch mode {
	case domain.RouteModeSafest, domain.RouteModeRiskiest:
		return float64(r.Danger)
	case domain.RouteModeQuickest:
		var logPose float64
		if isl, ok := islandsByID[toIsland]; ok && isl != nil {
			logPose = isl.LogPoseHours
		}
		return r.TravelHours + logPose
	default: // fastest
		return r.Distance
	}
}

// enrichPath recorre el camino una vez y produce:
//   - los PathStep con nombres y métricas acumuladas hasta ese paso.
//   - los totales globales del camino (4 métricas).
//
// Reglas:
//   - DistanceSoFar:    suma de Distance de aristas recorridas.
//   - TimeSoFar:        suma de TravelHours + LogPoseHours de islas intermedias
//     (excluye origen y destino).
//   - WorstDangerSoFar: max(Danger) de aristas recorridas.
//   - BestDangerSoFar:  min(Danger) de aristas recorridas.
//   - CostSoFar:        métrica del modo activo (legacy).
//
// Para path de 1 elemento (origen == destino), todas las métricas son 0.
func enrichPath(
	path []string,
	mode domain.RouteMode,
	routesByEdge map[string]domain.Route,
	islandsByID map[string]*domain.Island,
) ([]domain.PathStep, pathTotals) {
	steps := make([]domain.PathStep, len(path))
	totals := pathTotals{}

	for i, id := range path {
		name := id
		if isl, ok := islandsByID[id]; ok && isl != nil {
			name = isl.Name
		}

		if i > 0 {
			r, ok := routesByEdge[edgeKey(path[i-1], id)]
			if ok {
				totals.distance += r.Distance
				totals.time += r.TravelHours
				if r.Danger > totals.worstDanger {
					totals.worstDanger = r.Danger
				}
				if totals.bestDanger == 0 || r.Danger < totals.bestDanger {
					totals.bestDanger = r.Danger
				}
			}
			// LogPose de la isla intermedia anterior (path[i-1] cuando i-1 > 0).
			// El origen (path[0]) y el destino (path[len-1]) NO suman LogPose.
			if i-1 > 0 {
				if isl, ok := islandsByID[path[i-1]]; ok && isl != nil {
					totals.time += isl.LogPoseHours
				}
			}
		}

		steps[i] = domain.PathStep{
			IslandID:         id,
			IslandName:       name,
			DistanceSoFar:    totals.distance,
			TimeSoFar:        totals.time,
			WorstDangerSoFar: totals.worstDanger,
			BestDangerSoFar:  totals.bestDanger,
			CostSoFar:        costSoFarForMode(mode, totals),
		}
	}

	return steps, totals
}

// costSoFarForMode mapea las métricas acumuladas al campo legacy CostSoFar
// según el modo activo (mantiene compat con clientes que aún leen ese campo).
func costSoFarForMode(mode domain.RouteMode, t pathTotals) float64 {
	switch mode {
	case domain.RouteModeQuickest:
		return t.time
	case domain.RouteModeSafest:
		return float64(t.worstDanger)
	case domain.RouteModeRiskiest:
		return float64(t.bestDanger)
	default: // fastest
		return t.distance
	}
}

// totalCostForMode hace lo mismo que costSoFarForMode pero a nivel de respuesta global.
func totalCostForMode(mode domain.RouteMode, t pathTotals) float64 {
	return costSoFarForMode(mode, t)
}
