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
}
