package domain

// Island representa una localización del mundo de One Piece.
// Las coordenadas X/Y usan el sistema del mapa mundial:
//   - X: 0 (Oeste/East Blue) → 10000 (Este/New World final)
//   - Y: 0 (Sur/South Blue)  → 5000  (Norte/North Blue)
//   - Grand Line corre horizontalmente en Y ≈ 2500
type Island struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Region            string   `json:"region"`
	Description       string   `json:"description"`
	X                 float64  `json:"x"`
	Y                 float64  `json:"y"`
	NotableCharacters []string `json:"notable_characters,omitempty"`
}
