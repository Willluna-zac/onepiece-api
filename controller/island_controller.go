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

// GetAllIslands — GET /api/islands
func (c *IslandController) GetAllIslands(w http.ResponseWriter, r *http.Request) {
	islands, err := c.usecase.GetAllIslands(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetAllIslands: %v", err)
		sendError(w, http.StatusInternalServerError, "Error interno al obtener islas")
		return
	}
	sendJSON(w, http.StatusOK, islands)
}

// GetIslandByID — GET /api/islands/{id}
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

// GetIslandsByRegion — GET /api/islands/region/{region}
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

// GetNearestIsland — GET /api/islands/nearest?x=4300&y=2500
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
