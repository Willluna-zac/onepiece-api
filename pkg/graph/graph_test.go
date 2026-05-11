package graph

import (
	"math"
	"testing"
)

func TestDijkstra_SimpleGraph(t *testing.T) {
	g := New()
	g.AddBidirectionalEdge("A", "B", 1)
	g.AddBidirectionalEdge("B", "C", 2)
	g.AddBidirectionalEdge("A", "C", 5)

	result := g.Dijkstra("A", "C")
	if !result.Found {
		t.Fatal("expected path to be found")
	}
	if result.TotalCost != 3 {
		t.Errorf("expected cost 3, got %f", result.TotalCost)
	}
	// Camino optimo: A -> B -> C (costo 3) en vez de A -> C (costo 5)
	expected := []string{"A", "B", "C"}
	if len(result.Path) != len(expected) {
		t.Fatalf("expected path %v, got %v", expected, result.Path)
	}
	for i := range expected {
		if result.Path[i] != expected[i] {
			t.Errorf("path[%d] = %s, expected %s", i, result.Path[i], expected[i])
		}
	}
}

func TestDijkstra_DirectedGraph(t *testing.T) {
	// Simula Reverse Mountain (solo ida) y Skypiea (solo ida)
	g := New()
	g.AddEdge("loguetown", "reverse-mountain", 200)
	g.AddEdge("reverse-mountain", "whiskey-peak", 566)
	// No hay arista whiskey-peak -> reverse-mountain

	result := g.Dijkstra("whiskey-peak", "loguetown")
	if result.Found {
		t.Error("expected no path from whiskey-peak back to loguetown")
	}

	// Pero en la direccion correcta si funciona
	result = g.Dijkstra("loguetown", "whiskey-peak")
	if !result.Found {
		t.Fatal("expected path from loguetown to whiskey-peak")
	}
	if result.TotalCost != 766 {
		t.Errorf("expected cost 766, got %f", result.TotalCost)
	}
}

func TestDijkstra_Unreachable(t *testing.T) {
	g := New()
	g.AddNode("A")
	g.AddNode("B") // sin aristas

	result := g.Dijkstra("A", "B")
	if result.Found {
		t.Error("expected no path between disconnected nodes")
	}
}

func TestDijkstra_SameNode(t *testing.T) {
	g := New()
	g.AddNode("A")

	result := g.Dijkstra("A", "A")
	if !result.Found {
		t.Fatal("expected path to self")
	}
	if result.TotalCost != 0 {
		t.Errorf("expected cost 0, got %f", result.TotalCost)
	}
	if len(result.Path) != 1 || result.Path[0] != "A" {
		t.Errorf("expected path [A], got %v", result.Path)
	}
}

func TestDijkstra_NonexistentNode(t *testing.T) {
	g := New()
	g.AddNode("A")

	result := g.Dijkstra("A", "Z")
	if result.Found {
		t.Error("expected no path to nonexistent node")
	}

	result = g.Dijkstra("Z", "A")
	if result.Found {
		t.Error("expected no path from nonexistent node")
	}
}

func TestDijkstraAll(t *testing.T) {
	g := New()
	g.AddBidirectionalEdge("A", "B", 1)
	g.AddBidirectionalEdge("B", "C", 2)
	g.AddNode("D") // isla aislada

	dist := g.DijkstraAll("A")

	if dist["A"] != 0 {
		t.Errorf("dist to self should be 0, got %f", dist["A"])
	}
	if dist["B"] != 1 {
		t.Errorf("dist to B should be 1, got %f", dist["B"])
	}
	if dist["C"] != 3 {
		t.Errorf("dist to C should be 3, got %f", dist["C"])
	}
	if !math.IsInf(dist["D"], 1) {
		t.Errorf("dist to disconnected D should be Inf, got %f", dist["D"])
	}
}

func TestDijkstraAll_NonexistentSource(t *testing.T) {
	g := New()
	g.AddNode("A")

	dist := g.DijkstraAll("Z")
	if !math.IsInf(dist["A"], 1) {
		t.Errorf("all distances should be Inf for nonexistent source, got %f", dist["A"])
	}
}

func TestGraph_EdgeCount(t *testing.T) {
	g := New()
	g.AddBidirectionalEdge("A", "B", 1) // 2 aristas dirigidas
	g.AddEdge("B", "C", 2)              // 1 arista dirigida

	if g.EdgeCount() != 3 {
		t.Errorf("expected 3 directed edges, got %d", g.EdgeCount())
	}
}

func TestGraph_AddNodeIdempotent(t *testing.T) {
	g := New()
	g.AddNode("A")
	g.AddNode("A")
	g.AddNode("A")

	if len(g.Nodes()) != 1 {
		t.Errorf("expected 1 node after duplicate adds, got %d", len(g.Nodes()))
	}
}

func TestGraph_HasNode(t *testing.T) {
	g := New()
	g.AddNode("A")

	if !g.HasNode("A") {
		t.Error("expected HasNode(A) = true")
	}
	if g.HasNode("B") {
		t.Error("expected HasNode(B) = false")
	}
}

func TestGraph_Neighbors(t *testing.T) {
	g := New()
	g.AddEdge("A", "B", 1)
	g.AddEdge("A", "C", 2)

	neighbors := g.Neighbors("A")
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}

	// Nodo sin aristas salientes
	neighbors = g.Neighbors("B")
	if len(neighbors) != 0 {
		t.Errorf("expected 0 neighbors for B, got %d", len(neighbors))
	}

	// Nodo inexistente
	neighbors = g.Neighbors("Z")
	if neighbors != nil {
		t.Errorf("expected nil neighbors for nonexistent node, got %v", neighbors)
	}
}

// ---------------------------------------------------------------------------
// DijkstraBottleneck (minimax / maximin)
// ---------------------------------------------------------------------------

// bottleneckGrid construye el grafo de referencia usado en los tests:
//
//	     2
//	A ──────── B
//	│          │
//	5          1
//	│          │
//	D ──────── C
//	     3
//
// Caminos A → C:
//   - A → B → C: tramos [2, 1]
//   - A → D → C: tramos [5, 3]
//
// Bajo cada metrica el ganador difiere:
//   - Suma (Dijkstra clasico): A-B-C = 3   (gana frente a 8)
//   - Minimax: A-B-C max=2     (gana frente a max=5)
//   - Maximin: A-D-C min=3     (gana frente a min=1)
func bottleneckGrid() *Graph {
	g := New()
	g.AddBidirectionalEdge("A", "B", 2)
	g.AddBidirectionalEdge("A", "D", 5)
	g.AddBidirectionalEdge("B", "C", 1)
	g.AddBidirectionalEdge("C", "D", 3)
	return g
}

func assertPath(t *testing.T, got PathResult, wantCost float64, wantPath []string) {
	t.Helper()
	if !got.Found {
		t.Fatal("expected path to be found")
	}
	if got.TotalCost != wantCost {
		t.Errorf("expected cost %v, got %v", wantCost, got.TotalCost)
	}
	if len(got.Path) != len(wantPath) {
		t.Fatalf("expected path %v, got %v", wantPath, got.Path)
	}
	for i := range wantPath {
		if got.Path[i] != wantPath[i] {
			t.Errorf("path[%d] = %s, expected %s", i, got.Path[i], wantPath[i])
		}
	}
}

func TestDijkstraBottleneck_Min_PrefersSafestPath(t *testing.T) {
	g := bottleneckGrid()

	// Minimax A → C: queremos minimizar el peor tramo.
	// A-B-C tiene peor tramo 2; A-D-C tiene peor tramo 5. Gana A-B-C.
	got := g.DijkstraBottleneck("A", "C", BottleneckMin)
	assertPath(t, got, 2, []string{"A", "B", "C"})
}

func TestDijkstraBottleneck_Max_PrefersRiskiestPath(t *testing.T) {
	g := bottleneckGrid()

	// Maximin A → C: queremos maximizar el mejor tramo (sin tramos tranquilos).
	// A-B-C tiene mejor tramo 1; A-D-C tiene mejor tramo 3. Gana A-D-C.
	got := g.DijkstraBottleneck("A", "C", BottleneckMax)
	assertPath(t, got, 3, []string{"A", "D", "C"})
}

func TestDijkstraBottleneck_AllThreeModesDiverge(t *testing.T) {
	// Demuestra que el mismo grafo produce 3 rutas distintas bajo cada metrica.
	g := bottleneckGrid()

	classic := g.Dijkstra("A", "C")
	if classic.TotalCost != 3 {
		t.Errorf("clasico: esperado costo 3, got %v", classic.TotalCost)
	}

	safe := g.DijkstraBottleneck("A", "C", BottleneckMin)
	risky := g.DijkstraBottleneck("A", "C", BottleneckMax)

	// El clasico y el seguro escogen el mismo path en este grafo (ambos A-B-C),
	// pero su TotalCost difiere semanticamente: 3 (suma) vs 2 (peor tramo).
	if safe.TotalCost != 2 {
		t.Errorf("safest: esperado peor tramo 2, got %v", safe.TotalCost)
	}
	// El peligroso debe escoger el path opuesto.
	if risky.Path[1] != "D" {
		t.Errorf("riskiest: esperado pasar por D, got path %v", risky.Path)
	}
}

func TestDijkstraBottleneck_SameNode(t *testing.T) {
	g := bottleneckGrid()

	for _, mode := range []BottleneckMode{BottleneckMin, BottleneckMax} {
		got := g.DijkstraBottleneck("A", "A", mode)
		if !got.Found {
			t.Errorf("mode=%v: expected path to self", mode)
		}
		if got.TotalCost != 0 {
			t.Errorf("mode=%v: expected cost 0, got %v", mode, got.TotalCost)
		}
		if len(got.Path) != 1 || got.Path[0] != "A" {
			t.Errorf("mode=%v: expected path [A], got %v", mode, got.Path)
		}
	}
}

func TestDijkstraBottleneck_Unreachable(t *testing.T) {
	g := New()
	g.AddNode("A")
	g.AddNode("B") // sin aristas

	for _, mode := range []BottleneckMode{BottleneckMin, BottleneckMax} {
		got := g.DijkstraBottleneck("A", "B", mode)
		if got.Found {
			t.Errorf("mode=%v: expected no path between disconnected nodes", mode)
		}
	}
}

func TestDijkstraBottleneck_NonexistentNode(t *testing.T) {
	g := bottleneckGrid()

	for _, mode := range []BottleneckMode{BottleneckMin, BottleneckMax} {
		if got := g.DijkstraBottleneck("A", "Z", mode); got.Found {
			t.Errorf("mode=%v: expected no path to nonexistent target", mode)
		}
		if got := g.DijkstraBottleneck("Z", "A", mode); got.Found {
			t.Errorf("mode=%v: expected no path from nonexistent source", mode)
		}
	}
}

func TestDijkstraBottleneck_DirectedGraph(t *testing.T) {
	// Verifica que respeta la direccionalidad de las aristas.
	g := New()
	g.AddEdge("A", "B", 4)
	g.AddEdge("B", "C", 2)

	// Hay camino A→C pero no C→A.
	got := g.DijkstraBottleneck("A", "C", BottleneckMin)
	assertPath(t, got, 4, []string{"A", "B", "C"})

	got = g.DijkstraBottleneck("C", "A", BottleneckMin)
	if got.Found {
		t.Error("expected no reverse path in directed graph")
	}
}
