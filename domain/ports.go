package domain

import "context"

// CharacterRepository define el contrato que cualquier implementación
// de persistencia debe cumplir. Esto permite mockear en tests
// sin depender de Firestore real.
type CharacterRepository interface {
	CreateCharacter(ctx context.Context, character *Character) error
	GetAllCharacters(ctx context.Context) ([]Character, error)
	GetCharacterByID(ctx context.Context, id string) (*Character, error)
	UpdateCharacter(ctx context.Context, character *Character) error
	DeleteCharacter(ctx context.Context, id string) error

	// SearchByName hace una query prefix en Firestore sobre el campo "name"
	// (case-sensitive). Es O(k) donde k = resultados, no O(N) total.
	SearchByName(ctx context.Context, name string) ([]Character, error)

	// GetWithDevilFruit consulta la colección "devilfruits" directamente
	// y luego carga solo los personajes que tienen fruta. O(M) donde M << N.
	GetWithDevilFruit(ctx context.Context) ([]Character, error)
}

// RouteRepository define el contrato de persistencia para rutas marítimas.
type RouteRepository interface {
	GetAll(ctx context.Context) ([]Route, error)
	GetByIsland(ctx context.Context, islandID string) ([]Route, error)
}

// RouteUseCase define la lógica de negocio para rutas marítimas y navegación.
type RouteUseCase interface {
	GetAllRoutes(ctx context.Context) ([]Route, error)
	GetRoutesFromIsland(ctx context.Context, islandID string) ([]Route, error)
	FindShortestPath(ctx context.Context, fromID, toID string, mode RouteMode) (*ShortestPathResponse, error)
	FindReachableIslands(ctx context.Context, fromID string, maxCost float64) (*ReachableResponse, error)
	GetGraphStats(ctx context.Context) (*GraphStats, error)
}
