package repository

import (
	"context"
	"errors"
	"strings"

	"onepiece-api/domain"
	"onepiece-api/pkg/quadtree"
)

// IslandRepository almacena las islas en memoria e indexa su posición
// en un Quadtree para consultas de vecino más cercano en O(log n).
type IslandRepository struct {
	islands map[string]*domain.Island
	qt      *quadtree.Quadtree
}

// NewIslandRepository crea el repositorio y lo puebla con las 30 islas
// canónicas del mundo de One Piece, luego construye el Quadtree.
func NewIslandRepository() *IslandRepository {
	r := &IslandRepository{
		islands: make(map[string]*domain.Island),
		// Límites del mapa: X 0-10000, Y 0-5000
		qt: quadtree.New(quadtree.Bounds{MinX: 0, MinY: 0, MaxX: 10000, MaxY: 5000}, 4),
	}
	r.seed()
	return r
}

// GetAll retorna todas las islas ordenadas por ID.
func (r *IslandRepository) GetAll(_ context.Context) ([]*domain.Island, error) {
	result := make([]*domain.Island, 0, len(r.islands))
	for _, island := range r.islands {
		result = append(result, island)
	}
	return result, nil
}

// GetByID retorna una isla por su ID.
func (r *IslandRepository) GetByID(_ context.Context, id string) (*domain.Island, error) {
	island, ok := r.islands[id]
	if !ok {
		return nil, errors.New("isla no encontrada")
	}
	return island, nil
}

// GetByRegion retorna todas las islas de una región (case-insensitive).
func (r *IslandRepository) GetByRegion(_ context.Context, region string) ([]*domain.Island, error) {
	var result []*domain.Island
	term := strings.ToLower(region)
	for _, island := range r.islands {
		if strings.Contains(strings.ToLower(island.Region), term) {
			result = append(result, island)
		}
	}
	return result, nil
}

// GetNearest retorna la isla más cercana al punto (x, y) usando el Quadtree.
func (r *IslandRepository) GetNearest(_ context.Context, x, y float64) (*domain.Island, error) {
	nearest := r.qt.QueryNearest(x, y)
	if nearest == nil {
		return nil, errors.New("no hay islas registradas")
	}
	return nearest.Data.(*domain.Island), nil
}

// ---------------------------------------------------------------------------
// Datos de las 30 islas del mundo One Piece
// ---------------------------------------------------------------------------

func (r *IslandRepository) seed() {
	islands := []*domain.Island{
		{
			ID: "windmill-village", Name: "Windmill Village", Region: "East Blue",
			Description:       "Pueblo costero del Reino Goa donde Luffy creció y comenzó su viaje pirata.",
			X: 500, Y: 1800, NotableCharacters: []string{"Luffy", "Shanks", "Coby"},
		},
		{
			ID: "shells-town", Name: "Shells Town", Region: "East Blue",
			Description:       "Base marina donde Luffy recluta a Zoro de su ejecución en la jaula.",
			X: 700, Y: 2200, NotableCharacters: []string{"Zoro", "Coby", "Morgan"},
		},
		{
			ID: "orange-town", Name: "Orange Town", Region: "East Blue",
			Description:       "Primera villa atacada por los Straw Hats, donde se encuentran con Buggy el Payaso.",
			X: 900, Y: 2100, NotableCharacters: []string{"Buggy", "Nami"},
		},
		{
			ID: "syrup-village", Name: "Syrup Village", Region: "East Blue",
			Description:       "Tranquilo pueblo donde Usopp vive y se une a la tripulación del Sombrero de Paja.",
			X: 1100, Y: 2000, NotableCharacters: []string{"Usopp", "Kaya"},
		},
		{
			ID: "baratie", Name: "Baratie", Region: "East Blue",
			Description:       "Famoso restaurante flotante donde trabaja Sanji, que sueña con encontrar el All Blue.",
			X: 1400, Y: 2300, NotableCharacters: []string{"Sanji", "Zeff", "Gin"},
		},
		{
			ID: "arlong-park", Name: "Arlong Park", Region: "East Blue",
			Description:       "Fortaleza del tiburón humano Arlong, donde Nami estuvo esclavizada durante años.",
			X: 1600, Y: 1900, NotableCharacters: []string{"Nami", "Arlong"},
		},
		{
			ID: "loguetown", Name: "Loguetown", Region: "East Blue",
			Description:       "La Ciudad del Principio y el Fin: lugar de nacimiento y ejecución de Gol D. Roger.",
			X: 1900, Y: 2500, NotableCharacters: []string{"Luffy", "Buggy", "Smoker"},
		},
		{
			ID: "reverse-mountain", Name: "Reverse Mountain", Region: "Grand Line",
			Description:       "Entrada a la Grand Line: los barcos suben la corriente y se lanzan hacia el mar interior.",
			X: 2100, Y: 2500, NotableCharacters: []string{"Crocus", "Laboon"},
		},
		{
			ID: "whiskey-peak", Name: "Whiskey Peak", Region: "Grand Line",
			Description:       "Pueblo desértico donde los cazarrecompensas se disfrazan de anfitriones para atacar piratas.",
			X: 2500, Y: 2100, NotableCharacters: []string{"Igaram", "Zoro"},
		},
		{
			ID: "little-garden", Name: "Little Garden", Region: "Grand Line",
			Description:       "Isla prehistórica con criaturas gigantes donde dos gigantes llevan un duelo eterno.",
			X: 2800, Y: 2600, NotableCharacters: []string{"Dorry", "Brogy"},
		},
		{
			ID: "drum-island", Name: "Drum Island", Region: "Grand Line",
			Description:       "Reino nevado con el médico más habilidoso del mundo; donde Chopper se une a la tripulación.",
			X: 3000, Y: 2400, NotableCharacters: []string{"Chopper", "Dalton", "Wapol"},
		},
		{
			ID: "alabasta", Name: "Alabasta", Region: "Grand Line",
			Description:       "Vasto reino desértico al borde de la guerra civil, salvado por los Straw Hats de Crocodile.",
			X: 3200, Y: 2300, NotableCharacters: []string{"Vivi", "Crocodile", "Luffy"},
		},
		{
			ID: "jaya", Name: "Jaya", Region: "Grand Line",
			Description:       "Refugio de piratas donde Luffy se entera de Sky Island y enfrenta a Bellamy.",
			X: 3600, Y: 2200, NotableCharacters: []string{"Bellamy", "Luffy", "Blackbeard"},
		},
		{
			ID: "skypiea", Name: "Skypiea", Region: "Sky Islands",
			Description:       "Isla celeste flotando sobre las nubes, gobernada por el dios autoproclamado Enel.",
			X: 3800, Y: 4800, NotableCharacters: []string{"Enel", "Nami", "Zoro"},
		},
		{
			ID: "long-ring-long-land", Name: "Long Ring Long Land", Region: "Grand Line",
			Description:       "Isla alargada y peculiar donde los Straw Hats compiten en el Davy Back Fight contra Foxy.",
			X: 4000, Y: 2400, NotableCharacters: []string{"Foxy", "Luffy"},
		},
		{
			ID: "water-7", Name: "Water 7", Region: "Grand Line",
			Description:       "Ciudad flotante famosa por la construcción de barcos; donde Usopp abandona temporalmente la tripulación.",
			X: 4300, Y: 2700, NotableCharacters: []string{"Franky", "Iceburg", "Rob Lucci"},
		},
		{
			ID: "enies-lobby", Name: "Enies Lobby", Region: "Grand Line",
			Description:       "Fortaleza judicial del Gobierno Mundial donde los Straw Hats rescatan a Robin y le declaran la guerra.",
			X: 4600, Y: 2600, NotableCharacters: []string{"Rob Lucci", "Spandam", "Luffy"},
		},
		{
			ID: "thriller-bark", Name: "Thriller Bark", Region: "Grand Line",
			Description:       "Inmenso barco fantasma del Warlord Gecko Moriah; donde Brook se une a la tripulación.",
			X: 4800, Y: 2800, NotableCharacters: []string{"Gecko Moriah", "Brook", "Luffy"},
		},
		{
			ID: "sabaody-archipelago", Name: "Sabaody Archipelago", Region: "Grand Line",
			Description:       "Archipiélago de manglares burbujeantes donde la tripulación es separada por Bartholomew Kuma.",
			X: 4900, Y: 2500, NotableCharacters: []string{"Rayleigh", "Bartholomew Kuma", "Luffy"},
		},
		{
			ID: "amazon-lily", Name: "Amazon Lily", Region: "New World",
			Description:       "Isla de guerreras amazonas gobernada por Boa Hancock, donde Luffy aterriza tras la separación.",
			X: 5300, Y: 1200, NotableCharacters: []string{"Boa Hancock", "Luffy"},
		},
		{
			ID: "impel-down", Name: "Impel Down", Region: "Red Line",
			Description:       "La mayor prisión subterránea del mundo, asaltada por Luffy para rescatar a Ace.",
			X: 5000, Y: 1800, NotableCharacters: []string{"Magellan", "Luffy", "Ace"},
		},
		{
			ID: "marineford", Name: "Marineford", Region: "Red Line",
			Description:       "Cuartel general de la Marina donde se libra la mayor guerra del anime: la batalla de Marineford.",
			X: 5000, Y: 2400, NotableCharacters: []string{"Sengoku", "Whitebeard", "Luffy", "Ace"},
		},
		{
			ID: "fishman-island", Name: "Fish-Man Island", Region: "New World",
			Description:       "Isla submarina gobernada por el rey Neptuno; Luffy protege a los peces-hombre del racismo.",
			X: 5500, Y: 2300, NotableCharacters: []string{"Shirahoshi", "King Neptune", "Luffy"},
		},
		{
			ID: "punk-hazard", Name: "Punk Hazard", Region: "New World",
			Description:       "Isla devastada mitad fuego mitad hielo, base experimental del científico Vegapunk.",
			X: 5800, Y: 2600, NotableCharacters: []string{"Smoker", "Doflamingo", "Caesar Clown"},
		},
		{
			ID: "dressrosa", Name: "Dressrosa", Region: "New World",
			Description:       "Reino convertido en tierra de juguetes por Doflamingo, liberado tras la batalla del Coliseo.",
			X: 6200, Y: 2200, NotableCharacters: []string{"Doflamingo", "Rebecca", "Luffy"},
		},
		{
			ID: "zou", Name: "Zou", Region: "New World",
			Description:       "Isla viviente sobre el lomo de un elefante milenario, hogar de la tribu Mink.",
			X: 6600, Y: 2900, NotableCharacters: []string{"Inuarashi", "Nekomamushi", "Luffy"},
		},
		{
			ID: "whole-cake-island", Name: "Whole Cake Island", Region: "New World",
			Description:       "Isla-tarta sede del Imperio Totto Land de la Emperatriz Big Mom; donde Luffy rescata a Sanji.",
			X: 6800, Y: 3100, NotableCharacters: []string{"Big Mom", "Sanji", "Luffy"},
		},
		{
			ID: "wano", Name: "Wano Country", Region: "New World",
			Description:       "País de samuráis aislado del mundo bajo el yugo del Emperador Kaido, liberado por Luffy.",
			X: 7200, Y: 2700, NotableCharacters: []string{"Kaido", "Momonosuke", "Luffy", "Zoro"},
		},
		{
			ID: "elbaf", Name: "Elbaf", Region: "New World",
			Description:       "Legendaria tierra de los gigantes, soñada por Usopp desde niño.",
			X: 7600, Y: 4200, NotableCharacters: []string{"Hajrudin", "Luffy"},
		},
		{
			ID: "laugh-tale", Name: "Laugh Tale", Region: "New World",
			Description:       "La isla final de la Grand Line donde Gol D. Roger dejó el One Piece.",
			X: 8800, Y: 2500, NotableCharacters: []string{"Gol D. Roger", "Luffy"},
		},
	}

	for _, island := range islands {
		r.islands[island.ID] = island
		r.qt.Insert(quadtree.Point{X: island.X, Y: island.Y, Data: island})
	}
}
