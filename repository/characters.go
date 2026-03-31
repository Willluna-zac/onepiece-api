package repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"onepiece-api/domain"
)

// ---------------------------------------------------------------------------
// Estructuras internas para las colecciones normalizadas (dual-write)
// ---------------------------------------------------------------------------

type normalizedCharacter struct {
	ID              string    `firestore:"id"`
	Name            string    `firestore:"name"`
	Alias           string    `firestore:"alias"`
	Species         string    `firestore:"species"`
	Role            string    `firestore:"role"`
	FirstAppearance string    `firestore:"firstAppearance"`
	Notes           string    `firestore:"notes,omitempty"`
	MigratedAt      time.Time `firestore:"migratedAt"`
}

type normalizedDevilFruit struct {
	FruitID     string    `firestore:"fruit_id"`
	CharacterID string    `firestore:"character_id"`
	Name        string    `firestore:"name"`
	Type        string    `firestore:"type"`
	Description string    `firestore:"description,omitempty"`
	UpdatedAt   time.Time `firestore:"updatedAt"`
}

type normalizedHaki struct {
	ID          string    `firestore:"id"`
	CharacterID string    `firestore:"character_id"`
	HakiTypeID  string    `firestore:"haki_type_id"`
	Proficiency string    `firestore:"proficiency"`
	Awakened    bool      `firestore:"awakened"`
	Notes       string    `firestore:"notes,omitempty"`
	UpdatedAt   time.Time `firestore:"updatedAt"`
}

type normalizedAbility struct {
	ID          string    `firestore:"id"`
	CharacterID string    `firestore:"character_id"`
	Type        string    `firestore:"type"`
	Notes       string    `firestore:"notes,omitempty"`
	UpdatedAt   time.Time `firestore:"updatedAt"`
}

// hakiTypeIDs mapea nombre de haki → ID de documento en hakitypes
var hakiTypeIDs = map[string]string{
	"Armament":    "armament",
	"Observation": "observation",
	"Conqueror":   "conqueror",
}

// hakiTypeNames mapea ID de hakitypes → nombre legible (inverso de hakiTypeIDs)
var hakiTypeNames = map[string]string{
	"armament":    "Armament",
	"observation": "Observation",
	"conqueror":   "Conqueror",
}

// CharacterRepository maneja las operaciones de datos de personajes
type CharacterRepository struct {
	client     *firestore.Client
	collection string
}

// NewCharacterRepository crea una nueva instancia del repositorio
func NewCharacterRepository(client *firestore.Client) *CharacterRepository {
	// Nombre de la colección (puede configurarse con variable de entorno)
	collection := os.Getenv("FIRESTORE_COLLECTION")
	if collection == "" {
		collection = "characters" // Valor por defecto
	}

	return &CharacterRepository{
		client:     client,
		collection: collection,
	}
}

// CreateCharacter crea un nuevo personaje en Firestore (dual-write)
func (r *CharacterRepository) CreateCharacter(ctx context.Context, character *domain.Character) error {
	// Escritura en colección original
	_, err := r.client.Collection(r.collection).Doc(character.ID).Set(ctx, character)
	if err != nil {
		return err
	}

	// Dual-write: escribir también en colecciones normalizadas
	return r.writeNormalized(ctx, character)
}

// assembleCharacter construye un Character completo desde las colecciones normalizadas
func (r *CharacterRepository) assembleCharacter(ctx context.Context, base normalizedCharacter) (domain.Character, error) {
	char := domain.Character{
		ID:              base.ID,
		Name:            base.Name,
		Alias:           base.Alias,
		Species:         base.Species,
		Role:            base.Role,
		FirstAppearance: base.FirstAppearance,
		Notes:           base.Notes,
	}

	// Fruta del diablo
	fruitIter := r.client.Collection("devilfruits").Where("character_id", "==", base.ID).Documents(ctx)
	defer fruitIter.Stop()
	if fruitDoc, err := fruitIter.Next(); err == nil {
		var fruit normalizedDevilFruit
		if err := fruitDoc.DataTo(&fruit); err == nil {
			char.DevilFruit = &domain.DevilFruit{
				Name:        fruit.Name,
				Type:        fruit.Type,
				Description: fruit.Description,
			}
		}
	}

	// Haki
	hakiIter := r.client.Collection("character_haki").Where("character_id", "==", base.ID).Documents(ctx)
	defer hakiIter.Stop()
	for {
		doc, err := hakiIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		var h normalizedHaki
		if err := doc.DataTo(&h); err == nil {
			hakiName := hakiTypeNames[h.HakiTypeID]
			if hakiName == "" {
				hakiName = h.HakiTypeID
			}
			char.HakiAbilities = append(char.HakiAbilities, domain.HakiAbility{
				HakiType:    hakiName,
				Proficiency: h.Proficiency,
				Awakened:    h.Awakened,
				Notes:       h.Notes,
			})
		}
	}

	// Abilities
	abilityIter := r.client.Collection("abilities").Where("character_id", "==", base.ID).Documents(ctx)
	defer abilityIter.Stop()
	for {
		doc, err := abilityIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			break
		}
		var a normalizedAbility
		if err := doc.DataTo(&a); err == nil {
			char.Abilities = append(char.Abilities, domain.Ability{
				Type:  a.Type,
				Notes: a.Notes,
			})
		}
	}

	return char, nil
}

// GetAllCharacters obtiene todos los personajes desde las colecciones normalizadas
func (r *CharacterRepository) GetAllCharacters(ctx context.Context) ([]domain.Character, error) {
	var characters []domain.Character

	iter := r.client.Collection("characters_new").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var base normalizedCharacter
		if err := doc.DataTo(&base); err != nil {
			return nil, err
		}

		char, err := r.assembleCharacter(ctx, base)
		if err != nil {
			return nil, err
		}
		characters = append(characters, char)
	}

	return characters, nil
}

// GetCharacterByID obtiene un personaje por su ID desde las colecciones normalizadas
func (r *CharacterRepository) GetCharacterByID(ctx context.Context, id string) (*domain.Character, error) {
	doc, err := r.client.Collection("characters_new").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var base normalizedCharacter
	if err := doc.DataTo(&base); err != nil {
		return nil, err
	}

	char, err := r.assembleCharacter(ctx, base)
	if err != nil {
		return nil, err
	}

	return &char, nil
}

// UpdateCharacter actualiza un personaje existente (dual-write)
func (r *CharacterRepository) UpdateCharacter(ctx context.Context, character *domain.Character) error {
	// Escritura en colección original
	_, err := r.client.Collection(r.collection).Doc(character.ID).Set(ctx, character)
	if err != nil {
		return err
	}

	// Dual-write: sincronizar colecciones normalizadas
	return r.writeNormalized(ctx, character)
}

// DeleteCharacter elimina un personaje (dual-write)
func (r *CharacterRepository) DeleteCharacter(ctx context.Context, id string) error {
	// Eliminar en colección original
	_, err := r.client.Collection(r.collection).Doc(id).Delete(ctx)
	if err != nil {
		return err
	}

	// Dual-write: eliminar de colecciones normalizadas
	return r.deleteNormalized(ctx, id)
}

// ---------------------------------------------------------------------------
// Helpers de dual-write
// ---------------------------------------------------------------------------

// maxBatchSize es el límite seguro de operaciones por batch de Firestore (máx 500).
const maxBatchSize = 499

// commitInBatches divide las operaciones en lotes de máximo maxBatchSize
// y las commitea secuencialmente, respetando el límite de Firestore (BUG-1 fix).
func (r *CharacterRepository) commitInBatches(ctx context.Context, ops []func(*firestore.WriteBatch)) error {
	for i := 0; i < len(ops); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(ops) {
			end = len(ops)
		}
		batch := r.client.Batch()
		for _, op := range ops[i:end] {
			op(batch)
		}
		if _, err := batch.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// writeNormalized escribe un personaje en las colecciones normalizadas.
// Primero elimina haki y abilities existentes para evitar docs huérfanos en updates
// (BUG-2 fix), luego escribe los nuevos. Usa commitInBatches para el límite de 500 (BUG-1 fix).
func (r *CharacterRepository) writeNormalized(ctx context.Context, character *domain.Character) error {
	var ops []func(*firestore.WriteBatch)
	now := time.Now()

	// 1. characters_new (campos base sin relaciones embebidas)
	normChar := normalizedCharacter{
		ID:              character.ID,
		Name:            character.Name,
		Alias:           character.Alias,
		Species:         character.Species,
		Role:            character.Role,
		FirstAppearance: character.FirstAppearance,
		Notes:           character.Notes,
		MigratedAt:      now,
	}
	charRef := r.client.Collection("characters_new").Doc(character.ID)
	ops = append(ops, func(b *firestore.WriteBatch) { b.Set(charRef, normChar) })

	// 2. devilfruits: Set si existe, Delete si se quitó
	fruitID := fmt.Sprintf("%s_fruit", character.ID)
	fruitRef := r.client.Collection("devilfruits").Doc(fruitID)
	if character.DevilFruit != nil {
		normFruit := normalizedDevilFruit{
			FruitID:     fruitID,
			CharacterID: character.ID,
			Name:        character.DevilFruit.Name,
			Type:        character.DevilFruit.Type,
			Description: character.DevilFruit.Description,
			UpdatedAt:   now,
		}
		ops = append(ops, func(b *firestore.WriteBatch) { b.Set(fruitRef, normFruit) })
	} else {
		ops = append(ops, func(b *firestore.WriteBatch) { b.Delete(fruitRef) })
	}

	// 3. Eliminar haki existentes antes de escribir los nuevos (BUG-2 fix).
	//    Sin este paso, reducir de 3 a 2 haki deja el tercero huérfano para siempre.
	hakiIter := r.client.Collection("character_haki").Where("character_id", "==", character.ID).Documents(ctx)
	for {
		doc, err := hakiIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			hakiIter.Stop()
			return fmt.Errorf("error leyendo haki existente de %s: %w", character.ID, err)
		}
		ref := doc.Ref
		ops = append(ops, func(b *firestore.WriteBatch) { b.Delete(ref) })
	}
	hakiIter.Stop()

	// 4. Eliminar abilities existentes antes de escribir las nuevas (BUG-2 fix).
	abilityIter := r.client.Collection("abilities").Where("character_id", "==", character.ID).Documents(ctx)
	for {
		doc, err := abilityIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			abilityIter.Stop()
			return fmt.Errorf("error leyendo abilities existentes de %s: %w", character.ID, err)
		}
		ref := doc.Ref
		ops = append(ops, func(b *firestore.WriteBatch) { b.Delete(ref) })
	}
	abilityIter.Stop()

	// 5. Escribir nuevos haki
	for i, haki := range character.HakiAbilities {
		hakiID := fmt.Sprintf("%s_haki_%d", character.ID, i)
		hakiTypeID, ok := hakiTypeIDs[haki.HakiType]
		if !ok {
			hakiTypeID = "unknown"
		}
		normHaki := normalizedHaki{
			ID:          hakiID,
			CharacterID: character.ID,
			HakiTypeID:  hakiTypeID,
			Proficiency: haki.Proficiency,
			Awakened:    haki.Awakened,
			Notes:       haki.Notes,
			UpdatedAt:   now,
		}
		ref := r.client.Collection("character_haki").Doc(hakiID)
		ops = append(ops, func(b *firestore.WriteBatch) { b.Set(ref, normHaki) })
	}

	// 6. Escribir nuevas abilities
	for i, ability := range character.Abilities {
		abilityID := fmt.Sprintf("%s_ability_%d", character.ID, i)
		normAbility := normalizedAbility{
			ID:          abilityID,
			CharacterID: character.ID,
			Type:        ability.Type,
			Notes:       ability.Notes,
			UpdatedAt:   now,
		}
		ref := r.client.Collection("abilities").Doc(abilityID)
		ops = append(ops, func(b *firestore.WriteBatch) { b.Set(ref, normAbility) })
	}

	return r.commitInBatches(ctx, ops)
}

// deleteNormalized elimina todos los documentos normalizados de un personaje.
// Usa commitInBatches para respetar el límite de 500 ops de Firestore (BUG-1 fix).
func (r *CharacterRepository) deleteNormalized(ctx context.Context, id string) error {
	var ops []func(*firestore.WriteBatch)

	// characters_new
	charRef := r.client.Collection("characters_new").Doc(id)
	ops = append(ops, func(b *firestore.WriteBatch) { b.Delete(charRef) })

	// devilfruits
	fruitRef := r.client.Collection("devilfruits").Doc(fmt.Sprintf("%s_fruit", id))
	ops = append(ops, func(b *firestore.WriteBatch) { b.Delete(fruitRef) })

	// character_haki y abilities: query por character_id para no depender de IDs posicionales
	for _, coll := range []string{"character_haki", "abilities"} {
		iter := r.client.Collection(coll).Where("character_id", "==", id).Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				iter.Stop()
				return fmt.Errorf("error listando %s para delete de %s: %w", coll, id, err)
			}
			ref := doc.Ref
			ops = append(ops, func(b *firestore.WriteBatch) { b.Delete(ref) })
		}
		iter.Stop()
	}

	return r.commitInBatches(ctx, ops)
}
