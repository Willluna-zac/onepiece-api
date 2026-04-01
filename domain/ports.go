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
