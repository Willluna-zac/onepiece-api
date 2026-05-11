package graph

// Edge representa una arista ponderada en el grafo
type Edge struct {
	To     string  // ID del nodo destino
	Weight float64 // costo de la arista
}

// Graph implementa un grafo dirigido ponderado con adjacency list
type Graph struct {
	adjacency map[string][]Edge // nodo -> lista de aristas salientes
	nodes     map[string]bool   // set de nodos existentes
}

// New crea un grafo vacio
func New() *Graph {
	return &Graph{
		adjacency: make(map[string][]Edge),
		nodes:     make(map[string]bool),
	}
}

// AddNode agrega un nodo al grafo (idempotente)
func (g *Graph) AddNode(id string) {
	g.nodes[id] = true
	if _, exists := g.adjacency[id]; !exists {
		g.adjacency[id] = []Edge{}
	}
}

// AddEdge agrega una arista dirigida de from -> to con peso weight
func (g *Graph) AddEdge(from, to string, weight float64) {
	g.AddNode(from)
	g.AddNode(to)
	g.adjacency[from] = append(g.adjacency[from], Edge{To: to, Weight: weight})
}

// AddBidirectionalEdge agrega aristas en ambas direcciones
func (g *Graph) AddBidirectionalEdge(from, to string, weight float64) {
	g.AddEdge(from, to, weight)
	g.AddEdge(to, from, weight)
}

// Neighbors retorna las aristas salientes de un nodo
func (g *Graph) Neighbors(id string) []Edge {
	return g.adjacency[id]
}

// Nodes retorna todos los IDs de nodos
func (g *Graph) Nodes() []string {
	result := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		result = append(result, id)
	}
	return result
}

// HasNode verifica si un nodo existe
func (g *Graph) HasNode(id string) bool {
	return g.nodes[id]
}

// EdgeCount retorna el numero total de aristas dirigidas
func (g *Graph) EdgeCount() int {
	count := 0
	for _, edges := range g.adjacency {
		count += len(edges)
	}
	return count
}
