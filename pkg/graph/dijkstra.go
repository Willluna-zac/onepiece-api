package graph

import (
	"container/heap"
	"math"
)

// PathResult contiene el resultado de una busqueda de camino mas corto
type PathResult struct {
	Path      []string // secuencia de nodos desde origen hasta destino
	TotalCost float64  // costo total del camino
	Found     bool     // si se encontro un camino
}

// Dijkstra encuentra el camino mas corto entre source y target
func (g *Graph) Dijkstra(source, target string) PathResult {
	if !g.HasNode(source) || !g.HasNode(target) {
		return PathResult{Found: false}
	}

	if source == target {
		return PathResult{
			Path:      []string{source},
			TotalCost: 0,
			Found:     true,
		}
	}

	// Distancias minimas conocidas
	dist := make(map[string]float64, len(g.nodes))
	prev := make(map[string]string, len(g.nodes))
	for id := range g.nodes {
		dist[id] = math.Inf(1)
	}
	dist[source] = 0

	// Min-heap de prioridad
	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{node: source, cost: 0})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*pqItem)

		// Llegamos al destino
		if current.node == target {
			return PathResult{
				Path:      reconstructPath(prev, source, target),
				TotalCost: dist[target],
				Found:     true,
			}
		}

		// Si ya encontramos un camino mas corto, skip
		if current.cost > dist[current.node] {
			continue
		}

		// Explorar vecinos
		for _, edge := range g.Neighbors(current.node) {
			newDist := dist[current.node] + edge.Weight
			if newDist < dist[edge.To] {
				dist[edge.To] = newDist
				prev[edge.To] = current.node
				heap.Push(pq, &pqItem{node: edge.To, cost: newDist})
			}
		}
	}

	return PathResult{Found: false}
}

// DijkstraAll calcula las distancias mas cortas desde source a TODOS los nodos.
// Nodos inalcanzables quedan con distancia math.Inf(1).
func (g *Graph) DijkstraAll(source string) map[string]float64 {
	dist := make(map[string]float64, len(g.nodes))
	for id := range g.nodes {
		dist[id] = math.Inf(1)
	}

	if !g.HasNode(source) {
		return dist
	}

	dist[source] = 0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{node: source, cost: 0})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*pqItem)
		if current.cost > dist[current.node] {
			continue
		}
		for _, edge := range g.Neighbors(current.node) {
			newDist := dist[current.node] + edge.Weight
			if newDist < dist[edge.To] {
				dist[edge.To] = newDist
				heap.Push(pq, &pqItem{node: edge.To, cost: newDist})
			}
		}
	}

	return dist
}

// BottleneckMode controla la semantica de DijkstraBottleneck.
//
//   - BottleneckMin (minimax): minimiza el peso del peor tramo del camino.
//     Util para "ruta mas segura": el costo del camino es el peligro de su peor tramo.
//   - BottleneckMax (maximin): maximiza el peso del mejor tramo del camino.
//     Util para "ruta mas peligrosa": el costo del camino es el peligro de su mejor tramo,
//     y queremos que ese minimo sea lo mas alto posible (sin tramos tranquilos).
type BottleneckMode int

const (
	// BottleneckMin minimiza el maximo peso de arista en el camino.
	BottleneckMin BottleneckMode = iota
	// BottleneckMax maximiza el minimo peso de arista en el camino.
	BottleneckMax
)

// DijkstraBottleneck encuentra el camino optimo entre source y target bajo una
// metrica de cuello de botella en lugar de suma de pesos.
//
// A diferencia de Dijkstra clasico (que minimiza la suma de pesos), aqui:
//   - El "costo" de un camino es el max (BottleneckMin) o min (BottleneckMax)
//     del peso de sus aristas.
//   - El costo se compone con max/min en lugar de suma.
//
// El algoritmo sigue siendo greedy y sigue funcionando porque la propiedad de
// subestructura optima se mantiene: extender un camino solo puede empeorar
// (o mantener) su max, o empeorar (o mantener) su min.
func (g *Graph) DijkstraBottleneck(source, target string, mode BottleneckMode) PathResult {
	if !g.HasNode(source) || !g.HasNode(target) {
		return PathResult{Found: false}
	}

	if source == target {
		return PathResult{
			Path:      []string{source},
			TotalCost: 0,
			Found:     true,
		}
	}

	// Configuracion segun el modo.
	//   - initial: valor "peor que cualquiera" para nodos no visitados.
	//   - sourceInit: valor neutro al combinar con la primera arista.
	//   - better(a, b): true si a es estrictamente mejor que b.
	//   - combine(distU, w): nueva metrica al extender el camino con la arista w.
	var initial, sourceInit float64
	var better func(a, b float64) bool
	var combine func(distU, weight float64) float64

	if mode == BottleneckMin {
		// Minimax: empezamos en +Inf, mejoramos bajando.
		initial = math.Inf(1)
		sourceInit = 0
		better = func(a, b float64) bool { return a < b }
		combine = func(distU, weight float64) float64 { return math.Max(distU, weight) }
	} else {
		// Maximin: empezamos en -Inf, mejoramos subiendo.
		// Source arranca en +Inf para que el min con cualquier arista la "deje pasar".
		initial = math.Inf(-1)
		sourceInit = math.Inf(1)
		better = func(a, b float64) bool { return a > b }
		combine = func(distU, weight float64) float64 { return math.Min(distU, weight) }
	}

	dist := make(map[string]float64, len(g.nodes))
	prev := make(map[string]string, len(g.nodes))
	for id := range g.nodes {
		dist[id] = initial
	}
	dist[source] = sourceInit

	// Reutilizamos priorityQueue (min-heap). Para BottleneckMax negamos el costo
	// al empujar/leer asi el de mayor costo real sale primero.
	pushKey := func(c float64) float64 {
		if mode == BottleneckMax {
			return -c
		}
		return c
	}

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{node: source, cost: pushKey(sourceInit)})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*pqItem)
		realCost := current.cost
		if mode == BottleneckMax {
			realCost = -realCost
		}

		if current.node == target {
			return PathResult{
				Path:      reconstructPath(prev, source, target),
				TotalCost: dist[target],
				Found:     true,
			}
		}

		// Entrada obsoleta: ya encontramos algo mejor para este nodo.
		if better(dist[current.node], realCost) {
			continue
		}

		for _, edge := range g.Neighbors(current.node) {
			newDist := combine(dist[current.node], edge.Weight)
			if better(newDist, dist[edge.To]) {
				dist[edge.To] = newDist
				prev[edge.To] = current.node
				heap.Push(pq, &pqItem{node: edge.To, cost: pushKey(newDist)})
			}
		}
	}

	return PathResult{Found: false}
}

// reconstructPath reconstruye el camino desde source hasta target usando el mapa prev
func reconstructPath(prev map[string]string, source, target string) []string {
	path := []string{}
	current := target
	for current != "" {
		path = append([]string{current}, path...)
		if current == source {
			break
		}
		current = prev[current]
	}
	return path
}

// --- Priority Queue (min-heap) para Dijkstra ---

type pqItem struct {
	node  string
	cost  float64
	index int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].cost < pq[j].cost }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*pqItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[:n-1]
	return item
}
