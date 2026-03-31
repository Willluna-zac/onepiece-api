package domain

// Character representa un personaje de One Piece
type Character struct {
	ID              string        `firestore:"id" json:"id"`
	Name            string        `firestore:"name" json:"name"`
	Alias           string        `firestore:"alias" json:"alias"`
	Species         string        `firestore:"species" json:"species"`
	Role            string        `firestore:"role" json:"role"`
	FirstAppearance string        `firestore:"firstAppearance" json:"firstAppearance"`
	DevilFruit      *DevilFruit   `firestore:"devilFruit,omitempty" json:"devilFruit,omitempty"`
	HakiAbilities   []HakiAbility `firestore:"hakiAbilities,omitempty" json:"hakiAbilities,omitempty"`
	Abilities       []Ability     `firestore:"abilities,omitempty" json:"abilities,omitempty"`
	Notes           string        `firestore:"notes,omitempty" json:"notes,omitempty"`
}

// DevilFruit representa una fruta del diablo
type DevilFruit struct {
	Name        string `firestore:"name" json:"name"`
	Type        string `firestore:"type" json:"type"`
	Description string `firestore:"description,omitempty" json:"description,omitempty"`
}

// HakiAbility representa un tipo de Haki
type HakiAbility struct {
	HakiType    string `firestore:"hakiType" json:"hakiType"`
	Proficiency string `firestore:"proficiency" json:"proficiency"`
	Awakened    bool   `firestore:"awakened" json:"awakened"`
	Notes       string `firestore:"notes,omitempty" json:"notes,omitempty"`
}

// Ability representa una habilidad general
type Ability struct {
	Type  string `firestore:"type" json:"type"`
	Notes string `firestore:"notes,omitempty" json:"notes,omitempty"`
}
