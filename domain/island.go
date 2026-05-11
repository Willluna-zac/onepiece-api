package domain

// Island representa una localización del mundo de One Piece.
// Las coordenadas X/Y usan el sistema del mapa mundial:
//   - X: 0 (Oeste/East Blue) → 10000 (Este/New World final)
//   - Y: 0 (Sur/South Blue)  → 5000  (Norte/North Blue)
//   - Grand Line corre horizontalmente en Y ≈ 2500
//
// LogPoseHours representa el tiempo (en horas) que tarda el Log Pose
// en cargarse mientras la nave permanece en la isla antes de poder zarpar
// hacia la siguiente. Varía mucho segun la isla:
//   - East Blue / North Blue: 0 (no se usa Log Pose).
//   - Grand Line: 5–72 (Whisky Peak rapido, Drum Island ~3 dias).
//   - New World: 24–96 (Wano legendariamente lento).
//
// El LogPoseHours del origen y del destino final NO cuenta en el modo
// `quickest`: no se espera al zarpar de tu propia isla ni a la llegada.
type Island struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Region            string   `json:"region"`
	Description       string   `json:"description"`
	X                 float64  `json:"x"`
	Y                 float64  `json:"y"`
	LogPoseHours      float64  `json:"logPoseHours"`
	NotableCharacters []string `json:"notableCharacters,omitempty"`
}
