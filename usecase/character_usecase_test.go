package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"onepiece-api/domain"
	"onepiece-api/usecase"
)

// ---------------------------------------------------------------------------
// Mock del repositorio — implementa domain.CharacterRepository en memoria
// ---------------------------------------------------------------------------

type mockRepo struct {
	characters map[string]*domain.Character
	failOn     string // si coincide con el método, retorna error
}

func newMockRepo(initial ...*domain.Character) *mockRepo {
	m := &mockRepo{characters: make(map[string]*domain.Character)}
	for _, c := range initial {
		m.characters[c.ID] = c
	}
	return m
}

func (m *mockRepo) CreateCharacter(_ context.Context, c *domain.Character) error {
	if m.failOn == "CreateCharacter" {
		return errors.New("mock repo error")
	}
	m.characters[c.ID] = c
	return nil
}

func (m *mockRepo) GetAllCharacters(_ context.Context) ([]domain.Character, error) {
	if m.failOn == "GetAllCharacters" {
		return nil, errors.New("mock repo error")
	}
	result := make([]domain.Character, 0, len(m.characters))
	for _, c := range m.characters {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockRepo) GetCharacterByID(_ context.Context, id string) (*domain.Character, error) {
	if m.failOn == "GetCharacterByID" {
		return nil, errors.New("mock repo error")
	}
	c, ok := m.characters[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (m *mockRepo) UpdateCharacter(_ context.Context, c *domain.Character) error {
	if m.failOn == "UpdateCharacter" {
		return errors.New("mock repo error")
	}
	m.characters[c.ID] = c
	return nil
}

func (m *mockRepo) DeleteCharacter(_ context.Context, id string) error {
	if m.failOn == "DeleteCharacter" {
		return errors.New("mock repo error")
	}
	delete(m.characters, id)
	return nil
}

func (m *mockRepo) SearchByName(_ context.Context, name string) ([]domain.Character, error) {
	if m.failOn == "SearchByName" {
		return nil, errors.New("mock repo error")
	}
	var results []domain.Character
	term := strings.ToLower(name)
	for _, c := range m.characters {
		if strings.Contains(strings.ToLower(c.Name), term) || strings.Contains(strings.ToLower(c.Alias), term) {
			results = append(results, *c)
		}
	}
	return results, nil
}

func (m *mockRepo) GetWithDevilFruit(_ context.Context) ([]domain.Character, error) {
	if m.failOn == "GetWithDevilFruit" {
		return nil, errors.New("mock repo error")
	}
	var results []domain.Character
	for _, c := range m.characters {
		if c.DevilFruit != nil {
			results = append(results, *c)
		}
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func luffy() *domain.Character {
	return &domain.Character{
		ID:      "luffy-uuid",
		Name:    "Monkey D. Luffy",
		Alias:   "Mugiwara",
		Species: "Human",
		Role:    "Captain",
		DevilFruit: &domain.DevilFruit{
			Name: "Hito Hito no Mi",
			Type: "Mythical Zoan",
		},
		HakiAbilities: []domain.HakiAbility{
			{HakiType: "Conqueror", Proficiency: "Master", Awakened: true},
		},
	}
}

func zoro() *domain.Character {
	return &domain.Character{
		ID:      "zoro-uuid",
		Name:    "Roronoa Zoro",
		Alias:   "Pirate Hunter",
		Species: "Human",
		Role:    "Swordsman",
		HakiAbilities: []domain.HakiAbility{
			{HakiType: "Armament", Proficiency: "Advanced", Awakened: true},
		},
	}
}

// ---------------------------------------------------------------------------
// CreateCharacter
// ---------------------------------------------------------------------------

func TestCreateCharacter(t *testing.T) {
	tests := []struct {
		name      string
		input     *domain.Character
		repoFail  bool
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "crea personaje válido",
			input: &domain.Character{Name: "Nami", Species: "Human", Role: "Navigator"},
		},
		{
			name:      "falla si nombre vacío",
			input:     &domain.Character{Name: ""},
			wantErr:   true,
			errSubstr: "nombre es obligatorio",
		},
		{
			name:      "falla si nombre muy corto",
			input:     &domain.Character{Name: "A"},
			wantErr:   true,
			errSubstr: "al menos 2",
		},
		{
			name:      "falla si nombre muy largo",
			input:     &domain.Character{Name: strings.Repeat("x", 101)},
			wantErr:   true,
			errSubstr: "mas de 100",
		},
		{
			name:      "falla si tipo de fruta inválido",
			input:     &domain.Character{Name: "Nami", DevilFruit: &domain.DevilFruit{Name: "Test", Type: "Fake"}},
			wantErr:   true,
			errSubstr: "tipo de fruta",
		},
		{
			name:      "falla si proficiency de haki inválida",
			input:     &domain.Character{Name: "Nami", HakiAbilities: []domain.HakiAbility{{HakiType: "Armament", Proficiency: "God"}}},
			wantErr:   true,
			errSubstr: "nivel de haki",
		},
		{
			name:      "falla si el repo retorna error",
			input:     &domain.Character{Name: "Nami"},
			repoFail:  true,
			wantErr:   true,
			errSubstr: "mock repo error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			if tt.repoFail {
				repo.failOn = "CreateCharacter"
			}
			uc := usecase.NewCharacterUsecase(repo)
			err := uc.CreateCharacter(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("esperaba error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q no contiene %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("no esperaba error: %v", err)
			}
			// UUID debe haberse generado
			if tt.input.ID == "" {
				t.Error("el usecase debe generar un UUID")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetAllCharacters
// ---------------------------------------------------------------------------

func TestGetAllCharacters(t *testing.T) {
	t.Run("retorna todos los personajes", func(t *testing.T) {
		repo := newMockRepo(luffy(), zoro())
		uc := usecase.NewCharacterUsecase(repo)
		chars, err := uc.GetAllCharacters(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(chars) != 2 {
			t.Errorf("esperaba 2 personajes, got %d", len(chars))
		}
	})

	t.Run("retorna slice vacío sin personajes", func(t *testing.T) {
		repo := newMockRepo()
		uc := usecase.NewCharacterUsecase(repo)
		chars, err := uc.GetAllCharacters(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(chars) != 0 {
			t.Errorf("esperaba 0 personajes, got %d", len(chars))
		}
	})

	t.Run("propaga error del repo", func(t *testing.T) {
		repo := newMockRepo()
		repo.failOn = "GetAllCharacters"
		uc := usecase.NewCharacterUsecase(repo)
		_, err := uc.GetAllCharacters(context.Background())
		if err == nil {
			t.Fatal("esperaba error del repo")
		}
	})
}

// ---------------------------------------------------------------------------
// GetCharacterByID
// ---------------------------------------------------------------------------

func TestGetCharacterByID(t *testing.T) {
	repo := newMockRepo(luffy())
	uc := usecase.NewCharacterUsecase(repo)

	t.Run("retorna personaje existente", func(t *testing.T) {
		c, err := uc.GetCharacterByID(context.Background(), "luffy-uuid")
		if err != nil {
			t.Fatal(err)
		}
		if c.Name != "Monkey D. Luffy" {
			t.Errorf("esperaba Luffy, got %q", c.Name)
		}
	})

	t.Run("falla con ID vacío", func(t *testing.T) {
		_, err := uc.GetCharacterByID(context.Background(), "")
		if err == nil {
			t.Fatal("esperaba error con ID vacío")
		}
	})

	t.Run("falla con ID inexistente", func(t *testing.T) {
		_, err := uc.GetCharacterByID(context.Background(), "no-existe")
		if err == nil {
			t.Fatal("esperaba error con ID inexistente")
		}
	})
}

// ---------------------------------------------------------------------------
// UpdateCharacter
// ---------------------------------------------------------------------------

func TestUpdateCharacter(t *testing.T) {
	t.Run("actualiza personaje existente", func(t *testing.T) {
		repo := newMockRepo(luffy())
		uc := usecase.NewCharacterUsecase(repo)
		updated := &domain.Character{ID: "luffy-uuid", Name: "Luffy Gear 5", Species: "Human", Role: "Captain"}
		err := uc.UpdateCharacter(context.Background(), updated)
		if err != nil {
			t.Fatal(err)
		}
		c, _ := uc.GetCharacterByID(context.Background(), "luffy-uuid")
		if c.Name != "Luffy Gear 5" {
			t.Errorf("esperaba nombre actualizado, got %q", c.Name)
		}
	})

	t.Run("falla si personaje no existe", func(t *testing.T) {
		repo := newMockRepo()
		uc := usecase.NewCharacterUsecase(repo)
		err := uc.UpdateCharacter(context.Background(), &domain.Character{ID: "ghost", Name: "Fantasma"})
		if err == nil {
			t.Fatal("esperaba error")
		}
	})

	t.Run("falla con validación inválida", func(t *testing.T) {
		repo := newMockRepo(luffy())
		uc := usecase.NewCharacterUsecase(repo)
		err := uc.UpdateCharacter(context.Background(), &domain.Character{ID: "luffy-uuid", Name: ""})
		if err == nil {
			t.Fatal("esperaba error de validación")
		}
	})
}

// ---------------------------------------------------------------------------
// DeleteCharacter
// ---------------------------------------------------------------------------

func TestDeleteCharacter(t *testing.T) {
	t.Run("elimina personaje existente", func(t *testing.T) {
		repo := newMockRepo(luffy())
		uc := usecase.NewCharacterUsecase(repo)
		err := uc.DeleteCharacter(context.Background(), "luffy-uuid")
		if err != nil {
			t.Fatal(err)
		}
		_, err = uc.GetCharacterByID(context.Background(), "luffy-uuid")
		if err == nil {
			t.Fatal("el personaje debería haberse eliminado")
		}
	})

	t.Run("falla con ID vacío", func(t *testing.T) {
		repo := newMockRepo()
		uc := usecase.NewCharacterUsecase(repo)
		err := uc.DeleteCharacter(context.Background(), "")
		if err == nil {
			t.Fatal("esperaba error con ID vacío")
		}
	})

	t.Run("falla si personaje no existe", func(t *testing.T) {
		repo := newMockRepo()
		uc := usecase.NewCharacterUsecase(repo)
		err := uc.DeleteCharacter(context.Background(), "no-existe")
		if err == nil {
			t.Fatal("esperaba error")
		}
	})
}

// ---------------------------------------------------------------------------
// SearchByName
// ---------------------------------------------------------------------------

func TestSearchByName(t *testing.T) {
	repo := newMockRepo(luffy(), zoro())
	uc := usecase.NewCharacterUsecase(repo)

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantErr   bool
	}{
		{"encuentra por nombre exacto", "Luffy", 1, false},
		{"encuentra por alias", "Pirate Hunter", 1, false},
		{"búsqueda case-insensitive", "luffy", 1, false},
		{"búsqueda parcial", "Monkey", 1, false},
		{"sin resultados", "Shanks", 0, false},
		{"falla con query vacía", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := uc.SearchByName(context.Background(), tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatal("esperaba error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != tt.wantCount {
				t.Errorf("esperaba %d resultados, got %d", tt.wantCount, len(results))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetCharactersWithDevilFruit
// ---------------------------------------------------------------------------

func TestGetCharactersWithDevilFruit(t *testing.T) {
	t.Run("retorna solo personajes con fruta", func(t *testing.T) {
		repo := newMockRepo(luffy(), zoro()) // luffy tiene fruta, zoro no
		uc := usecase.NewCharacterUsecase(repo)
		results, err := uc.GetCharactersWithDevilFruit(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 {
			t.Errorf("esperaba 1 personaje con fruta, got %d", len(results))
		}
		if results[0].Name != "Monkey D. Luffy" {
			t.Errorf("esperaba Luffy, got %q", results[0].Name)
		}
	})

	t.Run("retorna vacío si nadie tiene fruta", func(t *testing.T) {
		repo := newMockRepo(zoro())
		uc := usecase.NewCharacterUsecase(repo)
		results, err := uc.GetCharactersWithDevilFruit(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Errorf("esperaba 0, got %d", len(results))
		}
	})
}
