package usecase_test

import (
	"context"
	"testing"

	"onepiece-api/domain"
	"onepiece-api/usecase"
)

// ---------------------------------------------------------------------------
// Mocks: RouteRepository + IslandRepository (subset que necesita el usecase)
// ---------------------------------------------------------------------------

type mockRouteRepo struct {
	routes []domain.Route
}

func (m *mockRouteRepo) GetAll(_ context.Context) ([]domain.Route, error) {
	return m.routes, nil
}

func (m *mockRouteRepo) GetByIsland(_ context.Context, _ string) ([]domain.Route, error) {
	return nil, nil
}

type mockIslandRepoForRoutes struct {
	islands map[string]*domain.Island
}

func (m *mockIslandRepoForRoutes) GetByID(_ context.Context, id string) (*domain.Island, error) {
	if isl, ok := m.islands[id]; ok {
		return isl, nil
	}
	return &domain.Island{ID: id, Name: id}, nil
}

func (m *mockIslandRepoForRoutes) GetAll(_ context.Context) ([]*domain.Island, error) {
	out := make([]*domain.Island, 0, len(m.islands))
	for _, isl := range m.islands {
		out = append(out, isl)
	}
	return out, nil
}

// buildRouteUsecase arma un RouteUsecase con el grafo de referencia:
//
//	     2 (d=2, t=10)
//	A ──────── B  (LogPose 100)
//	│                  │
//	5 (d=5, t=1)       1 (d=1, t=10)
//	│                  │
//	D (LogPose 0) ──── C
//	            3 (d=3, t=1)
//
// Distance == Danger en cada arista para que los tests fastest/safest/riskiest
// originales sigan funcionando. TravelHours y LogPose elegidos para que `quickest`
// prefiera A-D-C sobre A-B-C (la ruta rápida es la "larga en distancia" porque
// B tiene un LogPose enorme).
//
// Caminos A → C:
//   - A-B-C: dist [2,1] danger [2,1] travel [10,10] LogPose intermedio B=100
//   - A-D-C: dist [5,3] danger [5,3] travel [1,1]   LogPose intermedio D=0
//
// Esperado por modo:
//   - fastest:  A-B-C, totalDistance=3, totalTime=20+100=120
//   - safest:   A-B-C, worstDanger=2
//   - riskiest: A-D-C, bestDanger=3
//   - quickest: A-D-C, totalTime=2 (1+0+1)
func buildRouteUsecase() *usecase.RouteUsecase {
	rr := &mockRouteRepo{
		routes: []domain.Route{
			{ID: "ab", FromIsland: "A", ToIsland: "B", Distance: 2, Danger: 2, TravelHours: 10, Bidirectional: true},
			{ID: "ad", FromIsland: "A", ToIsland: "D", Distance: 5, Danger: 5, TravelHours: 1, Bidirectional: true},
			{ID: "bc", FromIsland: "B", ToIsland: "C", Distance: 1, Danger: 1, TravelHours: 10, Bidirectional: true},
			{ID: "cd", FromIsland: "C", ToIsland: "D", Distance: 3, Danger: 3, TravelHours: 1, Bidirectional: true},
		},
	}
	ir := &mockIslandRepoForRoutes{
		islands: map[string]*domain.Island{
			"A": {ID: "A", Name: "Isla A", LogPoseHours: 0},
			"B": {ID: "B", Name: "Isla B", LogPoseHours: 100},
			"C": {ID: "C", Name: "Isla C", LogPoseHours: 0},
			"D": {ID: "D", Name: "Isla D", LogPoseHours: 0},
		},
	}
	return usecase.NewRouteUseCase(rr, ir)
}

func pathIDs(resp *domain.ShortestPathResponse) []string {
	ids := make([]string, len(resp.Path))
	for i, s := range resp.Path {
		ids[i] = s.IslandID
	}
	return ids
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFindShortestPath_Fastest(t *testing.T) {
	uc := buildRouteUsecase()
	resp, err := uc.FindShortestPath(context.Background(), "A", "C", domain.RouteModeFastest)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Found {
		t.Fatal("expected path found")
	}
	if resp.Mode != domain.RouteModeFastest {
		t.Errorf("expected mode fastest, got %s", resp.Mode)
	}
	if resp.TotalCost != 3 {
		t.Errorf("expected total cost 3, got %v", resp.TotalCost)
	}
	if !equalSlice(pathIDs(resp), []string{"A", "B", "C"}) {
		t.Errorf("expected path A-B-C, got %v", pathIDs(resp))
	}
	// CostSoFar acumulado por suma
	if resp.Path[2].CostSoFar != 3 {
		t.Errorf("expected costSoFar=3 at C, got %v", resp.Path[2].CostSoFar)
	}
}

func TestFindShortestPath_Safest(t *testing.T) {
	uc := buildRouteUsecase()
	resp, err := uc.FindShortestPath(context.Background(), "A", "C", domain.RouteModeSafest)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Found {
		t.Fatal("expected path found")
	}
	if resp.Mode != domain.RouteModeSafest {
		t.Errorf("expected mode safest, got %s", resp.Mode)
	}
	// Peor tramo del camino A-B-C es 2 (max de 2,1)
	if resp.TotalCost != 2 {
		t.Errorf("expected total cost 2 (worst leg), got %v", resp.TotalCost)
	}
	if !equalSlice(pathIDs(resp), []string{"A", "B", "C"}) {
		t.Errorf("expected path A-B-C, got %v", pathIDs(resp))
	}
	// CostSoFar es el peor tramo hasta este paso (max acumulado)
	if resp.Path[2].CostSoFar != 2 {
		t.Errorf("expected costSoFar=2 at C, got %v", resp.Path[2].CostSoFar)
	}
}

func TestFindShortestPath_Riskiest(t *testing.T) {
	uc := buildRouteUsecase()
	resp, err := uc.FindShortestPath(context.Background(), "A", "C", domain.RouteModeRiskiest)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Found {
		t.Fatal("expected path found")
	}
	if resp.Mode != domain.RouteModeRiskiest {
		t.Errorf("expected mode riskiest, got %s", resp.Mode)
	}
	// Mejor tramo del camino A-D-C es 3 (min de 5,3); el otro camino tendria min=1.
	if resp.TotalCost != 3 {
		t.Errorf("expected total cost 3 (best leg), got %v", resp.TotalCost)
	}
	if !equalSlice(pathIDs(resp), []string{"A", "D", "C"}) {
		t.Errorf("expected path A-D-C, got %v", pathIDs(resp))
	}
}

func TestFindShortestPath_EmptyIDs(t *testing.T) {
	uc := buildRouteUsecase()
	if _, err := uc.FindShortestPath(context.Background(), "", "C", domain.RouteModeFastest); err == nil {
		t.Error("expected error for empty fromID")
	}
	if _, err := uc.FindShortestPath(context.Background(), "A", "", domain.RouteModeFastest); err == nil {
		t.Error("expected error for empty toID")
	}
}

func TestFindShortestPath_SameNode(t *testing.T) {
	uc := buildRouteUsecase()
	for _, mode := range []domain.RouteMode{
		domain.RouteModeFastest,
		domain.RouteModeQuickest,
		domain.RouteModeSafest,
		domain.RouteModeRiskiest,
	} {
		resp, err := uc.FindShortestPath(context.Background(), "A", "A", mode)
		if err != nil {
			t.Fatalf("mode=%s: %v", mode, err)
		}
		if !resp.Found {
			t.Errorf("mode=%s: expected found", mode)
		}
		if resp.TotalCost != 0 {
			t.Errorf("mode=%s: expected cost 0, got %v", mode, resp.TotalCost)
		}
		if resp.Hops != 0 {
			t.Errorf("mode=%s: expected hops 0, got %d", mode, resp.Hops)
		}
		if len(resp.Path) != 1 || resp.Path[0].IslandID != "A" {
			t.Errorf("mode=%s: expected path [A], got %v", mode, pathIDs(resp))
		}
	}
}

func TestFindShortestPath_Quickest(t *testing.T) {
	uc := buildRouteUsecase()
	resp, err := uc.FindShortestPath(context.Background(), "A", "C", domain.RouteModeQuickest)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Found {
		t.Fatal("expected path found")
	}
	if resp.Mode != domain.RouteModeQuickest {
		t.Errorf("expected mode quickest, got %s", resp.Mode)
	}
	// Quickest debe elegir A-D-C porque B tiene LogPose=100h. Costo en tiempo:
	//   A-B-C: 10 + 100 (LogPose B) + 10 = 120
	//   A-D-C: 1 + 0 (LogPose D) + 1 = 2  ← ganador
	if !equalSlice(pathIDs(resp), []string{"A", "D", "C"}) {
		t.Errorf("expected path A-D-C, got %v", pathIDs(resp))
	}
	if resp.TotalCost != 2 {
		t.Errorf("expected total cost 2 (totalTime), got %v", resp.TotalCost)
	}
	if resp.TotalTime != 2 {
		t.Errorf("expected totalTime=2, got %v", resp.TotalTime)
	}
	// Las otras métricas también deben venir llenas (calculadas sobre A-D-C):
	if resp.TotalDistance != 8 {
		t.Errorf("expected totalDistance=8 (5+3), got %v", resp.TotalDistance)
	}
	if resp.WorstDanger != 5 {
		t.Errorf("expected worstDanger=5, got %v", resp.WorstDanger)
	}
	if resp.BestDanger != 3 {
		t.Errorf("expected bestDanger=3, got %v", resp.BestDanger)
	}
}

func TestFindShortestPath_AlwaysReturnsAllMetrics(t *testing.T) {
	uc := buildRouteUsecase()
	for _, mode := range []domain.RouteMode{
		domain.RouteModeFastest,
		domain.RouteModeQuickest,
		domain.RouteModeSafest,
		domain.RouteModeRiskiest,
	} {
		resp, err := uc.FindShortestPath(context.Background(), "A", "C", mode)
		if err != nil {
			t.Fatalf("mode=%s: %v", mode, err)
		}
		if !resp.Found {
			t.Fatalf("mode=%s: expected found", mode)
		}
		// Las 4 métricas globales deben estar siempre presentes (>0).
		if resp.TotalDistance <= 0 {
			t.Errorf("mode=%s: totalDistance not populated (%v)", mode, resp.TotalDistance)
		}
		if resp.TotalTime <= 0 {
			t.Errorf("mode=%s: totalTime not populated (%v)", mode, resp.TotalTime)
		}
		if resp.WorstDanger <= 0 {
			t.Errorf("mode=%s: worstDanger not populated (%v)", mode, resp.WorstDanger)
		}
		if resp.BestDanger <= 0 {
			t.Errorf("mode=%s: bestDanger not populated (%v)", mode, resp.BestDanger)
		}
		// Y cada paso debe traer sus 4 acumulados.
		last := resp.Path[len(resp.Path)-1]
		if last.DistanceSoFar != resp.TotalDistance {
			t.Errorf("mode=%s: last DistanceSoFar (%v) != TotalDistance (%v)", mode, last.DistanceSoFar, resp.TotalDistance)
		}
		if last.TimeSoFar != resp.TotalTime {
			t.Errorf("mode=%s: last TimeSoFar (%v) != TotalTime (%v)", mode, last.TimeSoFar, resp.TotalTime)
		}
		if last.WorstDangerSoFar != resp.WorstDanger {
			t.Errorf("mode=%s: last WorstDangerSoFar (%v) != WorstDanger (%v)", mode, last.WorstDangerSoFar, resp.WorstDanger)
		}
		if last.BestDangerSoFar != resp.BestDanger {
			t.Errorf("mode=%s: last BestDangerSoFar (%v) != BestDanger (%v)", mode, last.BestDangerSoFar, resp.BestDanger)
		}
	}
}

// ---------------------------------------------------------------------------
// GetGraphStats
// ---------------------------------------------------------------------------

func TestGetGraphStats_BasicCounts(t *testing.T) {
	uc := buildRouteUsecase()
	stats, err := uc.GetGraphStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// 4 islas (A,B,C,D), 4 rutas, todas bidireccionales
	if stats.TotalIslands != 4 {
		t.Errorf("TotalIslands: want 4, got %d", stats.TotalIslands)
	}
	if stats.TotalRoutes != 4 {
		t.Errorf("TotalRoutes: want 4, got %d", stats.TotalRoutes)
	}
	if stats.BidirectionalCount != 4 {
		t.Errorf("BidirectionalCount: want 4, got %d", stats.BidirectionalCount)
	}
	if stats.IslandsWithLogPose != 1 { // sólo B (=100)
		t.Errorf("IslandsWithLogPose: want 1, got %d", stats.IslandsWithLogPose)
	}
	if stats.ConnectedComponents != 1 {
		t.Errorf("ConnectedComponents: want 1, got %d", stats.ConnectedComponents)
	}
	if stats.LargestComponent != 4 {
		t.Errorf("LargestComponent: want 4, got %d", stats.LargestComponent)
	}
	// Histograma de Danger (idx i = Danger i+1): rutas con danger 1,2,3,5
	want := [5]int{1, 1, 1, 0, 1}
	if stats.DangerHistogram != want {
		t.Errorf("DangerHistogram: want %v, got %v", want, stats.DangerHistogram)
	}
	// Promedios: dist (2+5+1+3)/4=2.75, time (10+1+10+1)/4=5.5, danger (2+5+1+3)/4=2.75
	if stats.AvgDistance != 2.75 {
		t.Errorf("AvgDistance: want 2.75, got %v", stats.AvgDistance)
	}
	if stats.AvgTravelHours != 5.5 {
		t.Errorf("AvgTravelHours: want 5.5, got %v", stats.AvgTravelHours)
	}
	if stats.AvgDanger != 2.75 {
		t.Errorf("AvgDanger: want 2.75, got %v", stats.AvgDanger)
	}
}

func TestGetGraphStats_DisconnectedGraph(t *testing.T) {
	rr := &mockRouteRepo{
		routes: []domain.Route{
			{ID: "ab", FromIsland: "A", ToIsland: "B", Distance: 1, Danger: 1, TravelHours: 1, Bidirectional: true},
			// C,D forman su propio componente
			{ID: "cd", FromIsland: "C", ToIsland: "D", Distance: 1, Danger: 1, TravelHours: 1, Bidirectional: true},
			// E aislada (la introduce el repo de islas)
		},
	}
	ir := &mockIslandRepoForRoutes{
		islands: map[string]*domain.Island{
			"A": {ID: "A", Name: "A"},
			"B": {ID: "B", Name: "B"},
			"C": {ID: "C", Name: "C"},
			"D": {ID: "D", Name: "D"},
			"E": {ID: "E", Name: "E"},
		},
	}
	uc := usecase.NewRouteUseCase(rr, ir)
	stats, err := uc.GetGraphStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.ConnectedComponents != 3 {
		t.Errorf("ConnectedComponents: want 3 (AB, CD, E), got %d", stats.ConnectedComponents)
	}
	if stats.LargestComponent != 2 {
		t.Errorf("LargestComponent: want 2, got %d", stats.LargestComponent)
	}
	if stats.TotalIslands != 5 {
		t.Errorf("TotalIslands: want 5, got %d", stats.TotalIslands)
	}
}
