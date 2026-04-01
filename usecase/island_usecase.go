package usecase

import (
	"context"
	"errors"
	"fmt"

	"onepiece-api/domain"
)

// IslandRepository define el contrato de persistencia para islas.
type IslandRepository interface {
	GetAll(ctx context.Context) ([]*domain.Island, error)
	GetByID(ctx context.Context, id string) (*domain.Island, error)
	GetByRegion(ctx context.Context, region string) ([]*domain.Island, error)
	GetNearest(ctx context.Context, x, y float64) (*domain.Island, error)
}

// IslandUsecase maneja la lógica de negocio del mapa de islas.
type IslandUsecase struct {
	repo IslandRepository
}

// NewIslandUsecase crea una nueva instancia del use case de islas.
func NewIslandUsecase(repo IslandRepository) *IslandUsecase {
	return &IslandUsecase{repo: repo}
}

// GetAllIslands retorna todas las islas registradas.
func (uc *IslandUsecase) GetAllIslands(ctx context.Context) ([]*domain.Island, error) {
	return uc.repo.GetAll(ctx)
}

// GetIslandByID retorna una isla por su identificador único.
func (uc *IslandUsecase) GetIslandByID(ctx context.Context, id string) (*domain.Island, error) {
	if id == "" {
		return nil, errors.New("el ID no puede estar vacío")
	}
	return uc.repo.GetByID(ctx, id)
}

// GetIslandsByRegion retorna todas las islas de una región.
func (uc *IslandUsecase) GetIslandsByRegion(ctx context.Context, region string) ([]*domain.Island, error) {
	if region == "" {
		return nil, errors.New("la región no puede estar vacía")
	}
	return uc.repo.GetByRegion(ctx, region)
}

// GetNearestIsland retorna la isla más cercana al punto (x, y) del mapa.
// Las coordenadas deben estar en el rango X: [0, 10000], Y: [0, 5000].
func (uc *IslandUsecase) GetNearestIsland(ctx context.Context, x, y float64) (*domain.Island, error) {
	if x < 0 || x > 10000 {
		return nil, fmt.Errorf("coordenada X %.1f fuera de rango [0, 10000]", x)
	}
	if y < 0 || y > 5000 {
		return nil, fmt.Errorf("coordenada Y %.1f fuera de rango [0, 5000]", y)
	}
	return uc.repo.GetNearest(ctx, x, y)
}
