package repository

import (
	"context"

	"onepiece-api/domain"
)

// RouteRepository almacena las rutas marítimas en memoria.
type RouteRepository struct {
	routes []domain.Route
}

// NewRouteRepository crea el repositorio y lo puebla con todas las rutas del mundo.
func NewRouteRepository() *RouteRepository {
	r := &RouteRepository{}
	r.seed()
	return r
}

// GetAll retorna todas las rutas marítimas.
func (r *RouteRepository) GetAll(_ context.Context) ([]domain.Route, error) {
	result := make([]domain.Route, len(r.routes))
	copy(result, r.routes)
	return result, nil
}

// GetByIsland retorna todas las rutas donde la isla dada es origen o destino.
func (r *RouteRepository) GetByIsland(_ context.Context, islandID string) ([]domain.Route, error) {
	var result []domain.Route
	for _, route := range r.routes {
		if route.FromIsland == islandID || route.ToIsland == islandID {
			result = append(result, route)
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Seed — rutas del mundo One Piece
// Weight = Distance * DangerMultiplier(Danger)
// ---------------------------------------------------------------------------

func (r *RouteRepository) seed() {
	routes := []domain.Route{

		// ── East Blue ──────────────────────────────────────────────────────
		{
			ID: "r01", FromIsland: "windmill-village", ToIsland: "shells-town",
			Distance: 200, Danger: 1, Bidirectional: true,
			Notes: "Primera etapa del viaje de Luffy",
		},
		{
			ID: "r02", FromIsland: "shells-town", ToIsland: "orange-town",
			Distance: 180, Danger: 1, Bidirectional: true,
			Notes: "Ruta costera del East Blue",
		},
		{
			ID: "r03", FromIsland: "orange-town", ToIsland: "syrup-village",
			Distance: 160, Danger: 1, Bidirectional: true,
			Notes: "Mar tranquilo del East Blue",
		},
		{
			ID: "r04", FromIsland: "syrup-village", ToIsland: "baratie",
			Distance: 250, Danger: 1, Bidirectional: true,
			Notes: "Hacia el restaurante flotante",
		},
		{
			ID: "r05", FromIsland: "baratie", ToIsland: "loguetown",
			Distance: 400, Danger: 2, Bidirectional: true,
			Notes: "Ciudad del Principio y el Fin",
		},
		{
			ID: "r06", FromIsland: "windmill-village", ToIsland: "kano-country",
			Distance: 600, Danger: 2, Bidirectional: true,
			Notes: "Ruta al North Blue cruzando el East Blue",
		},
		{
			ID: "r07", FromIsland: "baratie", ToIsland: "arlong-park",
			Distance: 224, Danger: 2, Bidirectional: true,
			Notes: "Nami va y vuelve entre Baratie y Arlong Park",
		},

		// ── Entrada a la Grand Line ─────────────────────────────────────────
		{
			ID: "r08", FromIsland: "loguetown", ToIsland: "reverse-mountain",
			Distance: 300, Danger: 3, Bidirectional: false,
			Notes: "Solo se puede entrar a la Grand Line, no volver por aquí",
		},
		{
			ID: "r09", FromIsland: "reverse-mountain", ToIsland: "whiskey-peak",
			Distance: 350, Danger: 3, Bidirectional: false,
			Notes: "Corriente del Log Pose hacia Paradise",
		},

		// ── Paradise ────────────────────────────────────────────────────────
		{
			ID: "r10", FromIsland: "whiskey-peak", ToIsland: "little-garden",
			Distance: 280, Danger: 2, Bidirectional: true,
			Notes: "Isla prehistórica de los gigantes",
		},
		{
			ID: "r11", FromIsland: "whiskey-peak", ToIsland: "drum-island",
			Distance: 400, Danger: 3, Bidirectional: true,
			Notes: "Ruta del Log Pose hacia el reino nevado",
		},
		{
			ID: "r12", FromIsland: "drum-island", ToIsland: "alabasta",
			Distance: 300, Danger: 2, Bidirectional: true,
			Notes: "Siguiente parada canónica",
		},
		{
			ID: "r13", FromIsland: "little-garden", ToIsland: "alabasta",
			Distance: 400, Danger: 2, Bidirectional: true,
			Notes: "Ruta directa evitando Drum Island",
		},
		{
			ID: "r14", FromIsland: "alabasta", ToIsland: "jaya",
			Distance: 420, Danger: 3, Bidirectional: true,
			Notes: "Hacia el refugio de piratas",
		},
		{
			ID: "r15", FromIsland: "jaya", ToIsland: "skypiea",
			Distance: 700, Danger: 4, Bidirectional: false,
			Notes: "Solo se sube a Sky Island desde Jaya con la corriente Knock Up Stream",
		},
		{
			ID: "r16", FromIsland: "skypiea", ToIsland: "jaya",
			Distance: 700, Danger: 3, Bidirectional: false,
			Notes: "Bajada desde Sky Island a través de la nube",
		},
		{
			ID: "r17", FromIsland: "jaya", ToIsland: "long-ring-long-land",
			Distance: 350, Danger: 2, Bidirectional: true,
			Notes: "Hacia el Davy Back Fight",
		},
		{
			ID: "r18", FromIsland: "long-ring-long-land", ToIsland: "water-7",
			Distance: 280, Danger: 2, Bidirectional: true,
			Notes: "Ciudad de los constructores de barcos",
		},
		{
			ID: "r19", FromIsland: "water-7", ToIsland: "enies-lobby",
			Distance: 120, Danger: 4, Bidirectional: true,
			Notes: "Fortaleza judicial muy próxima a Water 7",
		},
		{
			ID: "r20", FromIsland: "water-7", ToIsland: "thriller-bark",
			Distance: 350, Danger: 3, Bidirectional: true,
			Notes: "Barco fantasma de Gecko Moriah",
		},
		{
			ID: "r21", FromIsland: "thriller-bark", ToIsland: "sabaody-archipelago",
			Distance: 200, Danger: 3, Bidirectional: true,
			Notes: "Última parada antes de la Red Line",
		},
		{
			ID: "r22", FromIsland: "enies-lobby", ToIsland: "sabaody-archipelago",
			Distance: 500, Danger: 3, Bidirectional: true,
			Notes: "Ruta directa desde el juzgado al archipiélago",
		},
		{
			ID: "r23", FromIsland: "alabasta", ToIsland: "long-ring-long-land",
			Distance: 500, Danger: 2, Bidirectional: true,
			Notes: "Ruta alternativa por Paradise",
		},
		{
			ID: "r24", FromIsland: "whiskey-peak", ToIsland: "alabasta",
			Distance: 650, Danger: 2, Bidirectional: true,
			Notes: "Ruta alternativa del Log Pose",
		},

		// ── Red Line y zona submarina ────────────────────────────────────────
		{
			ID: "r25", FromIsland: "sabaody-archipelago", ToIsland: "fishman-island",
			Distance: 500, Danger: 4, Bidirectional: false,
			Notes: "Única ruta bajo la Red Line; no se puede regresar directamente",
		},
		{
			ID: "r26", FromIsland: "sabaody-archipelago", ToIsland: "marineford",
			Distance: 300, Danger: 4, Bidirectional: true,
			Notes: "Cuartel general de la Marina",
		},
		{
			ID: "r27", FromIsland: "marineford", ToIsland: "impel-down",
			Distance: 250, Danger: 4, Bidirectional: true,
			Notes: "La gran prisión submarina",
		},
		{
			ID: "r28", FromIsland: "impel-down", ToIsland: "amazon-lily",
			Distance: 350, Danger: 3, Bidirectional: true,
			Notes: "Isla de las guerreras amazonas",
		},
		{
			ID: "r29", FromIsland: "amazon-lily", ToIsland: "fishman-island",
			Distance: 450, Danger: 3, Bidirectional: true,
			Notes: "Ruta bajo el mar hasta la isla de los peces-hombre",
		},

		// ── New World ────────────────────────────────────────────────────────
		{
			ID: "r30", FromIsland: "fishman-island", ToIsland: "punk-hazard",
			Distance: 400, Danger: 4, Bidirectional: true,
			Notes: "Isla devastada del Nuevo Mundo",
		},
		{
			ID: "r31", FromIsland: "punk-hazard", ToIsland: "dressrosa",
			Distance: 350, Danger: 4, Bidirectional: true,
			Notes: "Reino de los juguetes de Doflamingo",
		},
		{
			ID: "r32", FromIsland: "dressrosa", ToIsland: "zou",
			Distance: 400, Danger: 3, Bidirectional: true,
			Notes: "Isla sobre el elefante milenario",
		},
		{
			ID: "r33", FromIsland: "zou", ToIsland: "whole-cake-island",
			Distance: 350, Danger: 4, Bidirectional: true,
			Notes: "Imperio Totto Land de Big Mom",
		},
		{
			ID: "r34", FromIsland: "zou", ToIsland: "wano",
			Distance: 500, Danger: 5, Bidirectional: true,
			Notes: "País de samuráis bajo Kaido",
		},
		{
			ID: "r35", FromIsland: "whole-cake-island", ToIsland: "wano",
			Distance: 400, Danger: 5, Bidirectional: true,
			Notes: "Alianza Big Mom–Kaido",
		},
		{
			ID: "r36", FromIsland: "wano", ToIsland: "elbaf",
			Distance: 600, Danger: 4, Bidirectional: true,
			Notes: "Tierra de los gigantes",
		},
		{
			ID: "r37", FromIsland: "elbaf", ToIsland: "laugh-tale",
			Distance: 900, Danger: 5, Bidirectional: false,
			Notes: "Solo ida: la isla final del One Piece",
		},
		{
			ID: "r38", FromIsland: "wano", ToIsland: "laugh-tale",
			Distance: 1100, Danger: 5, Bidirectional: false,
			Notes: "Ruta directa desde Wano a Laugh Tale",
		},

		// ── Conexiones especiales ────────────────────────────────────────────
		{
			ID: "r39", FromIsland: "banaro-island", ToIsland: "alabasta",
			Distance: 380, Danger: 3, Bidirectional: true,
			Notes: "Isla donde se enfrentaron Shanks y Barba Negra",
		},
		{
			ID: "r40", FromIsland: "kano-country", ToIsland: "reverse-mountain",
			Distance: 700, Danger: 3, Bidirectional: true,
			Notes: "Ruta desde North Blue hacia la Grand Line",
		},

		// ── Rutas alternativas (v3) — añaden divergencia entre modos ───────
		// Cada par siguiente ya existe con UN camino; estas aristas crean un
		// segundo camino con trade-off real (más larga pero más segura, o
		// más corta pero más peligrosa).
		{
			ID: "r41", FromIsland: "skypiea", ToIsland: "long-ring-long-land",
			Distance: 950, Danger: 1, Bidirectional: true,
			Notes: "Bajada controlada vía Octopus Balloons — ruta panorámica y segura",
		},
		{
			ID: "r42", FromIsland: "long-ring-long-land", ToIsland: "alabasta",
			Distance: 850, Danger: 2, Bidirectional: true,
			Notes: "Ruta sur del Grand Line — pasando por estepas tranquilas",
		},
		{
			ID: "r43", FromIsland: "sabaody-archipelago", ToIsland: "amazon-lily",
			Distance: 320, Danger: 5, Bidirectional: true,
			Notes: "Atajo por el Calm Belt — sin viento, lleno de Reyes del Mar",
		},
		{
			ID: "r44", FromIsland: "amazon-lily", ToIsland: "fishman-island",
			Distance: 280, Danger: 5, Bidirectional: true,
			Notes: "Continúa por Calm Belt — territorio prohibido",
		},
		{
			ID: "r45", FromIsland: "water-7", ToIsland: "thriller-bark",
			Distance: 1100, Danger: 4, Bidirectional: true,
			Notes: "Travesía nocturna directa — evita Enies Lobby pero atraviesa Florian Triangle",
		},
		{
			ID: "r46", FromIsland: "wano", ToIsland: "whole-cake-island",
			Distance: 1400, Danger: 4, Bidirectional: true,
			Notes: "Ruta directa Wano↔Totland — territorio Big Mom",
		},
		{
			ID: "r47", FromIsland: "drum-island", ToIsland: "alabasta",
			Distance: 600, Danger: 2, Bidirectional: true,
			Notes: "Atajo invierno→desierto — corrientes frías favorables",
		},
		{
			ID: "r48", FromIsland: "jaya", ToIsland: "water-7",
			Distance: 720, Danger: 3, Bidirectional: true,
			Notes: "Ruta directa post-Skypiea evitando Long Ring",
		},
	}

	// Calcular Weight (legacy) y TravelHours por ruta.
	//
	// TravelHours = Distance / velocidadRegional(islaOrigen). Velocidad expresada
	// en "unidades de Distance por hora", calibrada para que los modos divergan
	// visiblemente en pares con LogPoseHours alto:
	//   - East Blue / North Blue:    30  (mar tranquilo)
	//   - Grand Line / Red Line:     18  (corrientes erráticas, Log Pose obligatorio)
	//   - Sky Islands:               25  (vientos celestiales)
	//   - New World:                 12  (extremo)
	//
	// Algunas rutas icónicas se ajustan manualmente abajo (Knock Up Stream
	// rapidísimo, ruta submarina Sabaody→Fishman lenta, etc.).
	for i := range routes {
		routes[i].Weight = routes[i].Distance * domain.DangerMultiplier(routes[i].Danger)
		routes[i].TravelHours = routes[i].Distance / regionalVelocity(routes[i].FromIsland)
	}
	applyTravelHourOverrides(routes)

	r.routes = routes
}

// regionalVelocity retorna la velocidad de navegación característica de la
// región a la que pertenece la isla origen, usada para derivar TravelHours.
// Mapeado por ID de isla (las 32 del seed). Para islas no listadas, default 18.
func regionalVelocity(islandID string) float64 {
	switch islandID {
	// East Blue
	case "windmill-village", "shells-town", "orange-town", "syrup-village",
		"baratie", "arlong-park", "loguetown":
		return 30
	// North Blue
	case "kano-country":
		return 30
	// Sky Islands
	case "skypiea":
		return 25
	// New World
	case "amazon-lily", "fishman-island", "punk-hazard", "dressrosa",
		"zou", "whole-cake-island", "wano", "elbaf", "laugh-tale":
		return 12
	// Red Line
	case "impel-down", "marineford":
		return 18
	// Grand Line (default)
	default:
		return 18
	}
}

// applyTravelHourOverrides ajusta a mano TravelHours para rutas con tiempos
// canónicamente atípicos (la corriente vertical Knock Up Stream sube en horas;
// la ruta submarina Sabaody→Fishman tarda días por el coating).
func applyTravelHourOverrides(routes []domain.Route) {
	overrides := map[string]float64{
		"r08": 12,   // Loguetown → Reverse Mountain (corriente ayuda a entrar)
		"r15": 3.5,  // Jaya → Skypiea (Knock Up Stream, vertical y rápido)
		"r16": 35,   // Skypiea → Jaya (bajada por nube octopus, lento)
		"r25": 62.5, // Sabaody → Fishman Island (coating submarino, ~3 días)
		"r29": 50,   // Amazon Lily → Fishman Island (ruta submarina larga)
		"r37": 90,   // Elbaf → Laugh Tale (mística, navegación dificil)
		"r38": 110,  // Wano → Laugh Tale (ruta directa, mar agitado)
	}
	for i := range routes {
		if h, ok := overrides[routes[i].ID]; ok {
			routes[i].TravelHours = h
		}
	}
}
