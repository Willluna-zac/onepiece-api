// Audit examina la salud del seed de rutas/islas:
//   - Distribución de `danger` y `logPose`.
//   - Componentes conectados.
//   - Sample de pares aleatorios mostrando si los 4 modos convergen al mismo path
//     o producen rutas diferentes.
//
// Uso: go run ./migration/cmd/audit
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"onepiece-api/domain"
	"onepiece-api/repository"
	"onepiece-api/usecase"
)

func main() {
	ctx := context.Background()

	routeRepo := repository.NewRouteRepository()
	islandRepo := repository.NewIslandRepository()
	uc := usecase.NewRouteUseCase(routeRepo, islandRepo)

	routes, _ := routeRepo.GetAll(ctx)
	islands, _ := islandRepo.GetAll(ctx)

	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Println("                    GRAPH AUDIT REPORT                    ")
	fmt.Println("══════════════════════════════════════════════════════════")
	fmt.Printf("Islas: %d   |   Rutas: %d\n", len(islands), len(routes))
	fmt.Println()

	// ── 1) Distribución de Danger ─────────────────────────────────────────
	dangerCount := map[int]int{}
	for _, r := range routes {
		dangerCount[r.Danger]++
	}
	fmt.Println("─── Distribución de DANGER ─────────────────")
	for d := 1; d <= 5; d++ {
		c := dangerCount[d]
		pct := float64(c) / float64(len(routes)) * 100
		bar := strings.Repeat("█", c)
		fmt.Printf("  %d/5  %3d (%5.1f%%) %s\n", d, c, pct, bar)
	}
	fmt.Println()

	// ── 2) Distribución de LogPose en islas ──────────────────────────────
	logPoseBuckets := map[string]int{
		"sin Log Pose (0h)": 0,
		"corto  (1-12h)":    0,
		"medio  (13-48h)":   0,
		"largo  (49-96h)":   0,
		"épico  (>96h)":     0,
	}
	for _, isl := range islands {
		switch {
		case isl.LogPoseHours <= 0:
			logPoseBuckets["sin Log Pose (0h)"]++
		case isl.LogPoseHours <= 12:
			logPoseBuckets["corto  (1-12h)"]++
		case isl.LogPoseHours <= 48:
			logPoseBuckets["medio  (13-48h)"]++
		case isl.LogPoseHours <= 96:
			logPoseBuckets["largo  (49-96h)"]++
		default:
			logPoseBuckets["épico  (>96h)"]++
		}
	}
	fmt.Println("─── Distribución de LOG POSE ───────────────")
	keys := []string{"sin Log Pose (0h)", "corto  (1-12h)", "medio  (13-48h)", "largo  (49-96h)", "épico  (>96h)"}
	for _, k := range keys {
		fmt.Printf("  %-22s %3d\n", k, logPoseBuckets[k])
	}
	fmt.Println()

	// ── 3) Componentes conectados ────────────────────────────────────────
	adj := map[string][]string{}
	for _, r := range routes {
		adj[r.FromIsland] = append(adj[r.FromIsland], r.ToIsland)
		if r.Bidirectional {
			adj[r.ToIsland] = append(adj[r.ToIsland], r.FromIsland)
		}
	}
	visited := map[string]bool{}
	components := [][]string{}
	for _, isl := range islands {
		if visited[isl.ID] {
			continue
		}
		// BFS desde esta isla
		comp := []string{}
		queue := []string{isl.ID}
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			if visited[id] {
				continue
			}
			visited[id] = true
			comp = append(comp, id)
			queue = append(queue, adj[id]...)
		}
		components = append(components, comp)
	}
	fmt.Println("─── Componentes CONECTADOS ─────────────────")
	sort.Slice(components, func(i, j int) bool { return len(components[i]) > len(components[j]) })
	for i, c := range components {
		fmt.Printf("  #%d: %d islas\n", i+1, len(c))
		if len(c) <= 5 {
			fmt.Printf("       %v\n", c)
		}
	}
	fmt.Println()

	// ── 4) Sample de pares: convergencia de modos ────────────────────────
	fmt.Println("─── CONVERGENCIA de los 4 modos en pares random ─")
	fmt.Println("    (✅ = los 4 modos producen el MISMO path)")
	fmt.Println("    (🌟 = al menos un modo produce un path DIFERENTE)")
	fmt.Println()

	rng := rand.New(rand.NewSource(42))
	sampleN := 30
	converged := 0
	diverged := 0
	noRoute := 0
	modes := []domain.RouteMode{
		domain.RouteModeFastest, domain.RouteModeQuickest,
		domain.RouteModeSafest, domain.RouteModeRiskiest,
	}
	type sampleRow struct {
		from, to string
		ok       bool
		paths    [4]string
	}
	rows := []sampleRow{}

	for i := 0; i < sampleN; i++ {
		a := islands[rng.Intn(len(islands))]
		b := islands[rng.Intn(len(islands))]
		if a.ID == b.ID {
			i--
			continue
		}
		row := sampleRow{from: a.ID, to: b.ID, ok: true}
		var first string
		allEq := true
		for j, m := range modes {
			res, err := uc.FindShortestPath(ctx, a.ID, b.ID, m)
			if err != nil || !res.Found {
				row.ok = false
				break
			}
			ids := []string{}
			for _, s := range res.Path {
				ids = append(ids, s.IslandID)
			}
			joined := strings.Join(ids, ">")
			row.paths[j] = joined
			if j == 0 {
				first = joined
			} else if joined != first {
				allEq = false
			}
		}
		if !row.ok {
			noRoute++
		} else if allEq {
			converged++
		} else {
			diverged++
		}
		rows = append(rows, row)
	}

	fmt.Printf("  Convergen (mismo path):   %2d / %d (%.0f%%)\n", converged, sampleN, 100*float64(converged)/float64(sampleN))
	fmt.Printf("  Divergen (paths distintos): %2d / %d (%.0f%%)\n", diverged, sampleN, 100*float64(diverged)/float64(sampleN))
	fmt.Printf("  Sin ruta encontrada:      %2d / %d (%.0f%%)\n", noRoute, sampleN, 100*float64(noRoute)/float64(sampleN))
	fmt.Println()

	// Mostrar 5 ejemplos divergentes (los más interesantes para la demo)
	fmt.Println("─── Ejemplos DIVERGENTES (modos producen rutas distintas) ─")
	shown := 0
	for _, row := range rows {
		if !row.ok {
			continue
		}
		if row.paths[0] == row.paths[1] && row.paths[1] == row.paths[2] && row.paths[2] == row.paths[3] {
			continue
		}
		fmt.Printf("\n  %s → %s\n", row.from, row.to)
		labels := []string{"fastest ", "quickest", "safest  ", "riskiest"}
		for j, p := range row.paths {
			short := p
			if len(p) > 100 {
				short = p[:97] + "…"
			}
			fmt.Printf("    %s : %s\n", labels[j], short)
		}
		shown++
		if shown >= 5 {
			break
		}
	}
	if shown == 0 {
		fmt.Println("\n  ⚠ Ningún par random produjo modos divergentes en este sample.")
		fmt.Println("    Probable causa: faltan rutas alternativas con trade-offs.")
	}
	fmt.Println()
}
