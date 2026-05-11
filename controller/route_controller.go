package controller

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"onepiece-api/domain"
)

// routeUsecaseInterface define los métodos que el controller necesita del usecase de rutas.
type routeUsecaseInterface interface {
	GetAllRoutes(ctx context.Context) ([]domain.Route, error)
	GetRoutesFromIsland(ctx context.Context, islandID string) ([]domain.Route, error)
	FindShortestPath(ctx context.Context, fromID, toID string, mode domain.RouteMode) (*domain.ShortestPathResponse, error)
	FindReachableIslands(ctx context.Context, fromID string, maxCost float64) (*domain.ReachableResponse, error)
	GetGraphStats(ctx context.Context) (*domain.GraphStats, error)
}

// RouteController maneja los endpoints HTTP de rutas marítimas y navegación.
type RouteController struct {
	usecase routeUsecaseInterface
}

// NewRouteController crea una nueva instancia del controller de rutas.
func NewRouteController(uc routeUsecaseInterface) *RouteController {
	return &RouteController{usecase: uc}
}

// GetAllRoutes retorna todas las rutas marítimas registradas.
//
// @Summary      Listar rutas marítimas
// @Description  Retorna todas las rutas marítimas registradas en Firestore
// @Tags         routes
// @Produce      json
// @Success      200 {array}  domain.Route
// @Failure      500 {object} controller.ErrorResponse
// @Router       /api/routes [get]
func (c *RouteController) GetAllRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := c.usecase.GetAllRoutes(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetAllRoutes: %v", err)
		sendError(w, http.StatusInternalServerError, "Error interno al obtener rutas")
		return
	}
	sendJSON(w, http.StatusOK, routes)
}

// GetRoutesFromIsland retorna todas las rutas donde la isla dada es origen o destino.
//
// @Summary      Rutas desde/hacia una isla
// @Description  Retorna todas las rutas en las que la isla indicada participa como origen o destino
// @Tags         routes
// @Produce      json
// @Param        islandID path string true "ID de la isla"
// @Success      200 {array}  domain.Route
// @Failure      400 {object} controller.ErrorResponse
// @Failure      500 {object} controller.ErrorResponse
// @Router       /api/routes/from/{islandID} [get]
func (c *RouteController) GetRoutesFromIsland(w http.ResponseWriter, r *http.Request) {
	islandID := strings.TrimPrefix(r.URL.Path, "/api/routes/from/")
	if islandID == "" {
		sendError(w, http.StatusBadRequest, "El ID de isla es requerido")
		return
	}
	routes, err := c.usecase.GetRoutesFromIsland(r.Context(), islandID)
	if err != nil {
		log.Printf("[ERROR] GetRoutesFromIsland id=%s: %v", islandID, err)
		sendError(w, http.StatusInternalServerError, "Error interno al obtener rutas")
		return
	}
	sendJSON(w, http.StatusOK, routes)
}

// GetShortestPath calcula la ruta más corta entre dos islas usando Dijkstra.
// Query params: from, to, mode (opcional: fastest|quickest|safest|riskiest, default fastest).
//
// @Summary      Ruta más corta (Dijkstra)
// @Description  Calcula la ruta óptima entre dos islas según el modo elegido.
// @Description  - **fastest** (default): minimiza la suma de distancias.
// @Description  - **quickest**: minimiza el tiempo total (TravelHours de rutas + LogPoseHours de islas intermedias).
// @Description  - **safest**: minimiza el peligro del peor tramo (minimax sobre Danger).
// @Description  - **riskiest**: maximiza el peligro del mejor tramo (maximin sobre Danger).
// @Description  La respuesta SIEMPRE incluye las 4 métricas globales (`totalDistance`, `totalTime`, `worstDanger`, `bestDanger`) calculadas sobre el camino encontrado, independientemente del modo.
// @Description  El campo legacy `totalCost` refleja la métrica del modo activo (clientes nuevos deberían usar las 4 explícitas).
// @Tags         routes
// @Produce      json
// @Param        from query string true  "ID de la isla origen"
// @Param        to   query string true  "ID de la isla destino"
// @Param        mode query string false "Modo de búsqueda" Enums(fastest, quickest, safest, riskiest) default(fastest)
// @Success      200 {object} domain.ShortestPathResponse
// @Failure      400 {object} controller.ErrorResponse "Parámetros faltantes o modo inválido"
// @Failure      404 {object} controller.ErrorResponse "No existe ruta navegable"
// @Failure      500 {object} controller.ErrorResponse
// @Router       /api/routes/shortest [get]
func (c *RouteController) GetShortestPath(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		sendError(w, http.StatusBadRequest, "Los parámetros 'from' y 'to' son requeridos")
		return
	}

	mode, err := domain.ParseRouteMode(r.URL.Query().Get("mode"))
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.usecase.FindShortestPath(r.Context(), from, to, mode)
	if err != nil {
		log.Printf("[ERROR] GetShortestPath from=%s to=%s mode=%s: %v", from, to, mode, err)
		sendError(w, http.StatusInternalServerError, "Error interno al calcular ruta")
		return
	}

	if !result.Found {
		sendError(w, http.StatusNotFound, "No existe ruta navegable entre las islas indicadas")
		return
	}

	sendJSON(w, http.StatusOK, result)
}

// GetReachableIslands retorna todas las islas alcanzables desde un origen con costo <= maxCost.
// Query params: from, maxCost
//
// @Summary      Islas alcanzables
// @Description  Retorna todas las islas alcanzables desde el origen cuyo costo (suma de distancias) sea menor o igual a maxCost. Usa Dijkstra (modo fastest).
// @Tags         routes
// @Produce      json
// @Param        from    query string true "ID de la isla origen"
// @Param        maxCost query number true "Costo máximo permitido"
// @Success      200 {object} domain.ReachableResponse
// @Failure      400 {object} controller.ErrorResponse
// @Router       /api/routes/reachable [get]
func (c *RouteController) GetReachableIslands(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	maxCostStr := r.URL.Query().Get("maxCost")
	if from == "" || maxCostStr == "" {
		sendError(w, http.StatusBadRequest, "Los parámetros 'from' y 'maxCost' son requeridos")
		return
	}

	maxCost, err := strconv.ParseFloat(maxCostStr, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "El parámetro 'maxCost' debe ser un número")
		return
	}

	result, err := c.usecase.FindReachableIslands(r.Context(), from, maxCost)
	if err != nil {
		log.Printf("[ERROR] GetReachableIslands from=%s maxCost=%.1f: %v", from, maxCost, err)
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, result)
}

// GetGraphStats retorna metricas estructurales del grafo de rutas e islas.
//
// @Summary      Estadisticas del grafo
// @Description  Retorna metricas estructurales del grafo: totales, promedios (distancia, tiempo, danger), distribucion de Danger por nivel (1-5) y numero de componentes conectados.
// @Description  Util para visualizar la salud del seed y construir vistas de analisis.
// @Tags         routes
// @Produce      json
// @Success      200 {object} domain.GraphStats
// @Failure      500 {object} controller.ErrorResponse
// @Router       /api/routes/stats [get]
func (c *RouteController) GetGraphStats(w http.ResponseWriter, r *http.Request) {
	stats, err := c.usecase.GetGraphStats(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetGraphStats: %v", err)
		sendError(w, http.StatusInternalServerError, "Error interno al calcular stats del grafo")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=60")
	sendJSON(w, http.StatusOK, stats)
}
