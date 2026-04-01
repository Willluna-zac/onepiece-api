package controller

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"onepiece-api/domain"
)

// islandUsecaseInterface define los métodos que el controller necesita del usecase.
type islandUsecaseInterface interface {
	GetAllIslands(ctx context.Context) ([]*domain.Island, error)
	GetIslandByID(ctx context.Context, id string) (*domain.Island, error)
	GetIslandsByRegion(ctx context.Context, region string) ([]*domain.Island, error)
	GetNearestIsland(ctx context.Context, x, y float64) (*domain.Island, error)
}

// IslandController maneja los endpoints HTTP del mapa de islas.
type IslandController struct {
	usecase islandUsecaseInterface
}

// NewIslandController crea una nueva instancia del controller de islas.
func NewIslandController(uc islandUsecaseInterface) *IslandController {
	return &IslandController{usecase: uc}
}

// GetAllIslands godoc
// @Summary      Listar todas las islas
// @Description  Retorna las 30 islas canónicas del mundo de One Piece con coordenadas 2D
// @Tags         islands
// @Produce      json
// @Success      200  {array}   domain.Island
// @Router       /api/islands [get]
func (c *IslandController) GetAllIslands(w http.ResponseWriter, r *http.Request) {
	islands, err := c.usecase.GetAllIslands(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetAllIslands: %v", err)
		sendError(w, http.StatusInternalServerError, "Error interno al obtener islas")
		return
	}
	sendJSON(w, http.StatusOK, islands)
}

// GetIslandByID godoc
// @Summary      Obtener isla por ID
// @Description  Retorna el detalle de una isla dado su slug (ej: water-7, laugh-tale)
// @Tags         islands
// @Produce      json
// @Param        id   path      string  true  "Slug de la isla (ej: water-7)"
// @Success      200  {object}  domain.Island
// @Failure      404  {object}  map[string]string
// @Router       /api/islands/{id} [get]
func (c *IslandController) GetIslandByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/islands/")
	island, err := c.usecase.GetIslandByID(r.Context(), id)
	if err != nil {
		log.Printf("[ERROR] GetIslandByID id=%s: %v", id, err)
		sendError(w, http.StatusNotFound, "Isla no encontrada")
		return
	}
	sendJSON(w, http.StatusOK, island)
}

// GetIslandsByRegion godoc
// @Summary      Islas por región
// @Description  Filtra islas por región (East Blue, Grand Line, New World, Sky Islands, Red Line)
// @Tags         islands
// @Produce      json
// @Param        region  path      string  true  "Nombre de la región"
// @Success      200     {array}   domain.Island
// @Router       /api/islands/region/{region} [get]
func (c *IslandController) GetIslandsByRegion(w http.ResponseWriter, r *http.Request) {
	region := strings.TrimPrefix(r.URL.Path, "/api/islands/region/")
	if region == "" {
		sendError(w, http.StatusBadRequest, "La región es requerida")
		return
	}
	islands, err := c.usecase.GetIslandsByRegion(r.Context(), region)
	if err != nil {
		log.Printf("[ERROR] GetIslandsByRegion region=%s: %v", region, err)
		sendError(w, http.StatusInternalServerError, "Error interno al filtrar islas")
		return
	}
	sendJSON(w, http.StatusOK, islands)
}

// GetNearestIsland godoc
// @Summary      Isla más cercana (Quadtree)
// @Description  Dado un punto (x,y) del mapa, retorna la isla más cercana usando un índice Quadtree 2D. Coordenadas: X 0-10000 (Oeste→Este), Y 0-5000 (Sur→Norte). La Grand Line corre en Y≈2500.
// @Tags         islands
// @Produce      json
// @Param        x  query     number  true  "Coordenada X del mapa (0-10000)"  example(4300)
// @Param        y  query     number  true  "Coordenada Y del mapa (0-5000)"   example(2500)
// @Success      200  {object}  domain.Island
// @Failure      400  {object}  map[string]string
// @Router       /api/islands/nearest [get]
func (c *IslandController) GetNearestIsland(w http.ResponseWriter, r *http.Request) {
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")
	if xStr == "" || yStr == "" {
		sendError(w, http.StatusBadRequest, "Los parámetros 'x' e 'y' son requeridos")
		return
	}

	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "El parámetro 'x' debe ser un número")
		return
	}
	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "El parámetro 'y' debe ser un número")
		return
	}

	island, err := c.usecase.GetNearestIsland(r.Context(), x, y)
	if err != nil {
		log.Printf("[ERROR] GetNearestIsland x=%.1f y=%.1f: %v", x, y, err)
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}
	sendJSON(w, http.StatusOK, island)
}
