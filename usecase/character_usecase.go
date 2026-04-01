package usecase

import (
	"context"
	"errors"
	"strings"

	"onepiece-api/domain"

	"github.com/google/uuid"
)

// CharacterUsecase maneja la logica de negocio de personajes
type CharacterUsecase struct {
	repo domain.CharacterRepository
}

// NewCharacterUsecase crea una nueva instancia del use case.
// Acepta la interfaz domain.CharacterRepository para facilitar el testing con mocks.
func NewCharacterUsecase(repo domain.CharacterRepository) *CharacterUsecase {
	return &CharacterUsecase{repo: repo}
}

// CreateCharacter crea un nuevo personaje con validaciones
func (uc *CharacterUsecase) CreateCharacter(ctx context.Context, character *domain.Character) error {
	if err := uc.validateCharacter(character); err != nil {
		return err
	}
	character.ID = uuid.New().String()
	character.Name = strings.TrimSpace(character.Name)
	character.Alias = strings.TrimSpace(character.Alias)
	return uc.repo.CreateCharacter(ctx, character)
}

// GetAllCharacters obtiene todos los personajes
func (uc *CharacterUsecase) GetAllCharacters(ctx context.Context) ([]domain.Character, error) {
	return uc.repo.GetAllCharacters(ctx)
}

// GetCharacterByID obtiene un personaje por su ID
func (uc *CharacterUsecase) GetCharacterByID(ctx context.Context, id string) (*domain.Character, error) {
	if id == "" {
		return nil, errors.New("el ID no puede estar vacio")
	}
	return uc.repo.GetCharacterByID(ctx, id)
}

// UpdateCharacter actualiza un personaje existente
func (uc *CharacterUsecase) UpdateCharacter(ctx context.Context, character *domain.Character) error {
	existing, err := uc.repo.GetCharacterByID(ctx, character.ID)
	if err != nil {
		return errors.New("personaje no encontrado")
	}
	if existing == nil {
		return errors.New("personaje no existe")
	}
	if err := uc.validateCharacter(character); err != nil {
		return err
	}
	return uc.repo.UpdateCharacter(ctx, character)
}

// DeleteCharacter elimina un personaje
func (uc *CharacterUsecase) DeleteCharacter(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("el ID no puede estar vacio")
	}
	existing, err := uc.repo.GetCharacterByID(ctx, id)
	if err != nil {
		return errors.New("personaje no encontrado")
	}
	if existing == nil {
		return errors.New("personaje no existe")
	}
	return uc.repo.DeleteCharacter(ctx, id)
}

// SearchByName busca personajes por nombre delegando al repository.
// El repository usa una Firestore range query (prefix, case-sensitive).
func (uc *CharacterUsecase) SearchByName(ctx context.Context, name string) ([]domain.Character, error) {
	if name == "" {
		return nil, errors.New("el nombre no puede estar vacio")
	}
	return uc.repo.SearchByName(ctx, name)
}

// GetCharactersWithDevilFruit obtiene personajes con fruta del diablo
// delegando al repository, que consulta la colección devilfruits directamente.
func (uc *CharacterUsecase) GetCharactersWithDevilFruit(ctx context.Context) ([]domain.Character, error) {
	return uc.repo.GetWithDevilFruit(ctx)
}

// validateCharacter valida los datos de un personaje
func (uc *CharacterUsecase) validateCharacter(character *domain.Character) error {
	if character == nil {
		return errors.New("el personaje no puede ser nil")
	}
	if strings.TrimSpace(character.Name) == "" {
		return errors.New("el nombre es obligatorio")
	}
	if len(character.Name) < 2 {
		return errors.New("el nombre debe tener al menos 2 caracteres")
	}
	if len(character.Name) > 100 {
		return errors.New("el nombre no puede tener mas de 100 caracteres")
	}
	if character.DevilFruit != nil {
		validTypes := map[string]bool{
			"Paramecia": true, "Logia": true, "Zoan": true,
			"Mythical Zoan": true, "Ancient Zoan": true,
		}
		if !validTypes[character.DevilFruit.Type] {
			return errors.New("tipo de fruta del diablo invalido")
		}
	}
	if len(character.HakiAbilities) > 0 {
		validProf := map[string]bool{"Basic": true, "Advanced": true, "Master": true}
		for _, h := range character.HakiAbilities {
			if !validProf[h.Proficiency] {
				return errors.New("nivel de haki invalido (Basic, Advanced o Master)")
			}
		}
	}
	return nil
}
