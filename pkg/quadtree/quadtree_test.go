package quadtree_test

import (
	"math"
	"testing"

	"onepiece-api/pkg/quadtree"
)

var worldBounds = quadtree.Bounds{MinX: 0, MinY: 0, MaxX: 10000, MaxY: 5000}

func TestInsertAndSize(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)

	points := []quadtree.Point{
		{X: 100, Y: 200},
		{X: 500, Y: 300},
		{X: 9000, Y: 4000},
		{X: 5000, Y: 2500},
		{X: 3000, Y: 1000},
	}

	for _, p := range points {
		if !qt.Insert(p) {
			t.Errorf("Insert(%v) devolvió false, se esperaba true", p)
		}
	}

	if qt.Size() != len(points) {
		t.Errorf("Size() = %d, want %d", qt.Size(), len(points))
	}
}

func TestInsertFueraDeRango(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)
	p := quadtree.Point{X: -1, Y: 2500} // fuera del borde oeste
	if qt.Insert(p) {
		t.Error("Insert fuera de bounds debería retornar false")
	}
	if qt.Size() != 0 {
		t.Error("el árbol debería estar vacío")
	}
}

func TestSubdivision(t *testing.T) {
	// capacity=2 → con 5 puntos el nodo debe haberse subdividido
	qt := quadtree.New(worldBounds, 2)
	for i := 0; i < 5; i++ {
		qt.Insert(quadtree.Point{X: float64(i * 1000), Y: float64(i * 500)})
	}
	if qt.Size() != 5 {
		t.Errorf("Size() = %d, want 5", qt.Size())
	}
}

func TestQueryNearestVacio(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)
	if qt.QueryNearest(5000, 2500) != nil {
		t.Error("árbol vacío debe retornar nil")
	}
}

func TestQueryNearestUnSoloPunto(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)
	p := quadtree.Point{X: 3000, Y: 2500, Data: "Alabasta"}
	qt.Insert(p)

	result := qt.QueryNearest(3001, 2501)
	if result == nil {
		t.Fatal("se esperaba un resultado")
	}
	if result.Data != "Alabasta" {
		t.Errorf("got %v, want Alabasta", result.Data)
	}
}

func TestQueryNearestRetornaElMasCercano(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)

	islands := []quadtree.Point{
		{X: 500, Y: 1800, Data: "Windmill Village"}, // Este Blue, lejos
		{X: 4300, Y: 2700, Data: "Water 7"},          // Grand Line, cerca del query
		{X: 9000, Y: 2500, Data: "Laugh Tale"},        // New World, lejos
		{X: 4600, Y: 2600, Data: "Enies Lobby"},       // Grand Line, muy cerca
	}
	for _, p := range islands {
		qt.Insert(p)
	}

	// Buscar desde un punto muy cercano a Enies Lobby
	result := qt.QueryNearest(4650, 2620)
	if result == nil {
		t.Fatal("se esperaba un resultado")
	}
	if result.Data != "Enies Lobby" {
		t.Errorf("se esperaba Enies Lobby (más cercano), got %v", result.Data)
	}
}

func TestQueryNearestConMuchosPuntos(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)

	// Insertar puntos en cuadrícula: X cada 1000, Y cada 1000
	for i := 0; i <= 9; i++ {
		for j := 0; j <= 4; j++ {
			qt.Insert(quadtree.Point{X: float64(i * 1000), Y: float64(j * 1000)})
		}
	}

	// Buscar desde (5000, 2500) — punto entre (5000,2000) y (5000,3000), ambos a dist=500
	qx, qy := 5000.0, 2500.0
	result := qt.QueryNearest(qx, qy)
	if result == nil {
		t.Fatal("se esperaba un resultado")
	}

	// Verificar que la distancia del resultado es la mínima posible (500)
	gotDist := math.Sqrt((result.X-qx)*(result.X-qx) + (result.Y-qy)*(result.Y-qy))
	const wantDist = 500.0
	if math.Abs(gotDist-wantDist) > 1e-9 {
		t.Errorf("distancia al más cercano = %.2f, want %.2f (point: %.0f,%.0f)",
			gotDist, wantDist, result.X, result.Y)
	}
}

func TestQueryRange(t *testing.T) {
	qt := quadtree.New(worldBounds, 4)

	qt.Insert(quadtree.Point{X: 500, Y: 1800, Data: "Windmill"})  // East Blue
	qt.Insert(quadtree.Point{X: 4300, Y: 2700, Data: "Water 7"}) // Grand Line
	qt.Insert(quadtree.Point{X: 8800, Y: 2500, Data: "Laugh Tale"}) // New World

	// Rango que solo cubre East Blue
	eastBlue := quadtree.Bounds{MinX: 0, MinY: 0, MaxX: 2000, MaxY: 5000}
	results := qt.QueryRange(eastBlue)
	if len(results) != 1 {
		t.Errorf("QueryRange East Blue: got %d puntos, want 1", len(results))
	}
	if results[0].Data != "Windmill" {
		t.Errorf("got %v, want Windmill", results[0].Data)
	}
}
