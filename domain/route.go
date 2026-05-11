package domain

import (
	"errors"
	"strings"
)

// Route representa una ruta maritima entre dos islas.
//
// Cada ruta lleva las 3 metricas que el dominio necesita por arista:
//   - Distance:     unidades de distancia (escala arbitraria del mapa).
//   - TravelHours:  tiempo de navegacion del tramo. Independiente de Distance:
//     dos rutas con la misma Distance pueden tardar muy distinto
//     segun corrientes, calma chicha o tipo de mar (lore-friendly).
//   - Danger:       1-5, peligrosidad del tramo.
type Route struct {
	ID          string  `json:"id"`
	FromIsland  string  `json:"fromIsland"`
	ToIsland    string  `json:"toIsland"`
	Distance    float64 `json:"distance"`
	TravelHours float64 `json:"travelHours"`
	Danger      int     `json:"danger"` // 1-5
	// Deprecated: Weight es un peso precalculado heredado del primer Dijkstra
	// (Distance * DangerMultiplier(Danger)). Los modos de busqueda actuales
	// (fastest/safest/riskiest/quickest) calculan el peso dinamicamente desde
	// Distance, TravelHours o Danger segun el modo. Se conserva para no migrar
	// Firestore; no usar en codigo nuevo.
	Weight        float64 `json:"weight"`
	Bidirectional bool    `json:"bidirectional"`
	Notes         string  `json:"notes,omitempty"`
}

// RouteMode controla la semantica de la busqueda de ruta mas corta.
//
//   - RouteModeFastest:  minimiza la suma de distancias (Dijkstra clasico).
//   - RouteModeQuickest: minimiza el tiempo total = sum(TravelHours rutas) +
//     sum(LogPoseHours islas intermedias). Origen y destino
//     no aportan LogPoseHours (no se espera al zarpar ni al llegar).
//   - RouteModeSafest:   minimiza el peligro del peor tramo (minimax sobre Danger).
//   - RouteModeRiskiest: maximiza el peligro del mejor tramo (maximin sobre Danger).
type RouteMode string

const (
	RouteModeFastest  RouteMode = "fastest"
	RouteModeQuickest RouteMode = "quickest"
	RouteModeSafest   RouteMode = "safest"
	RouteModeRiskiest RouteMode = "riskiest"
)

// ErrInvalidRouteMode se retorna cuando el modo no corresponde a uno conocido.
var ErrInvalidRouteMode = errors.New("modo de ruta invalido (esperado: fastest, quickest, safest, riskiest)")

// ParseRouteMode convierte un string al RouteMode equivalente. Una cadena vacia
// se interpreta como RouteModeFastest (default seguro y retrocompatible).
func ParseRouteMode(s string) (RouteMode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return RouteModeFastest, nil
	case string(RouteModeFastest):
		return RouteModeFastest, nil
	case string(RouteModeQuickest):
		return RouteModeQuickest, nil
	case string(RouteModeSafest):
		return RouteModeSafest, nil
	case string(RouteModeRiskiest):
		return RouteModeRiskiest, nil
	default:
		return "", ErrInvalidRouteMode
	}
}

// ShortestPathResponse es la respuesta del endpoint de ruta mas corta.
//
// La respuesta es uniforme entre modos: SIEMPRE devuelve las 4 metricas globales
// (TotalDistance, TotalTime, WorstDanger, BestDanger) calculadas sobre el camino
// encontrado, independientemente del modo optimizado. El frontend resalta la
// metrica del modo activo y muestra las demas como informacion comparativa.
//
// TotalCost (legacy) mantiene la metrica del modo activo:
//   - fastest:  igual a TotalDistance.
//   - quickest: igual a TotalTime.
//   - safest:   igual a WorstDanger (peor tramo).
//   - riskiest: igual a BestDanger (mejor tramo).
//
// Clientes nuevos deberian usar las 4 metricas explicitas.
type ShortestPathResponse struct {
	From string    `json:"from"`
	To   string    `json:"to"`
	Mode RouteMode `json:"mode"`

	// Metricas globales del camino encontrado. Siempre presentes.
	TotalDistance float64 `json:"totalDistance"`
	TotalTime     float64 `json:"totalTime"`
	WorstDanger   int     `json:"worstDanger"`
	BestDanger    int     `json:"bestDanger"`

	// TotalCost (legacy) refleja la metrica del modo activo.
	TotalCost float64 `json:"totalCost"`

	Hops  int        `json:"hops"`
	Path  []PathStep `json:"path"`
	Found bool       `json:"found"`
}

// PathStep representa un paso individual en la ruta.
//
// Cada paso lleva las 4 metricas acumuladas hasta ese punto:
//   - DistanceSoFar:    suma de Distance de las aristas recorridas.
//   - TimeSoFar:        suma de TravelHours de aristas + LogPoseHours de islas
//     intermedias (excluye origen y destino).
//   - WorstDangerSoFar: max(Danger) de las aristas recorridas.
//   - BestDangerSoFar:  min(Danger) de las aristas recorridas.
//
// CostSoFar (legacy) refleja la metrica del modo activo (ver ShortestPathResponse).
type PathStep struct {
	IslandID   string `json:"islandId"`
	IslandName string `json:"islandName"`

	DistanceSoFar    float64 `json:"distanceSoFar"`
	TimeSoFar        float64 `json:"timeSoFar"`
	WorstDangerSoFar int     `json:"worstDangerSoFar"`
	BestDangerSoFar  int     `json:"bestDangerSoFar"`

	CostSoFar float64 `json:"costSoFar"`
}

// ReachableResponse para el endpoint /reachable
type ReachableResponse struct {
	From    string            `json:"from"`
	MaxCost float64           `json:"maxCost"`
	Islands []ReachableIsland `json:"islands"`
}

// ReachableIsland representa una isla alcanzable con su costo
type ReachableIsland struct {
	IslandID   string  `json:"islandId"`
	IslandName string  `json:"islandName"`
	Cost       float64 `json:"cost"`
}

// GraphStats agrupa metricas estructurales del grafo de islas y rutas.
// Se calcula sobre el snapshot actual del repositorio y se sirve cacheado
// brevemente en el controller.
type GraphStats struct {
	TotalIslands        int     `json:"totalIslands"`
	TotalRoutes         int     `json:"totalRoutes"`
	BidirectionalCount  int     `json:"bidirectionalCount"`
	IslandsWithLogPose  int     `json:"islandsWithLogPose"`
	ConnectedComponents int     `json:"connectedComponents"`
	LargestComponent    int     `json:"largestComponent"`
	AvgDistance         float64 `json:"avgDistance"`
	AvgTravelHours      float64 `json:"avgTravelHours"`
	AvgDanger           float64 `json:"avgDanger"`
	// DangerHistogram[i] = cantidad de rutas con Danger == i+1 (i en [0..4]).
	DangerHistogram [5]int `json:"dangerHistogram"`
}

// DangerMultiplier retorna el multiplicador de peligro para el calculo de Weight.
// Weight = Distance * DangerMultiplier(Danger)
func DangerMultiplier(danger int) float64 {
	switch danger {
	case 1:
		return 1.0
	case 2:
		return 1.3
	case 3:
		return 1.6
	case 4:
		return 2.0
	case 5:
		return 3.0
	default:
		return 1.0
	}
}
