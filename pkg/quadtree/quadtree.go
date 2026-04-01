// Package quadtree implementa un Quadtree 2D para indexación espacial eficiente.
// Permite insertar puntos y encontrar el vecino más cercano (nearest neighbor)
// a un punto dado — ideal para el mapa del mundo de One Piece.
package quadtree

import "math"

// Point es un punto en el espacio 2D con un payload arbitrario.
type Point struct {
	X, Y float64
	Data interface{}
}

// Bounds representa un rectángulo en el plano 2D.
type Bounds struct {
	MinX, MinY, MaxX, MaxY float64
}

// contains informa si el punto p está dentro de los límites (inclusive).
func (b Bounds) contains(p Point) bool {
	return p.X >= b.MinX && p.X <= b.MaxX &&
		p.Y >= b.MinY && p.Y <= b.MaxY
}

// minDistToPoint calcula la distancia mínima desde (x,y) al borde del rectángulo.
// Retorna 0 si el punto está dentro — se usa para podar ramas del árbol en QueryNearest.
func (b Bounds) minDistToPoint(x, y float64) float64 {
	dx := math.Max(b.MinX-x, math.Max(0, x-b.MaxX))
	dy := math.Max(b.MinY-y, math.Max(0, y-b.MaxY))
	return math.Sqrt(dx*dx + dy*dy)
}

// Quadtree divide el espacio 2D en 4 cuadrantes recursivamente.
// Cada nodo almacena hasta `capacity` puntos; al superarlo se subdivide.
//
//	+------+------+
//	|  NW  |  NE  |
//	+------+------+
//	|  SW  |  SE  |
//	+------+------+
type Quadtree struct {
	bounds   Bounds
	capacity int
	points   []Point
	divided  bool
	nw, ne   *Quadtree
	sw, se   *Quadtree
}

// New crea un Quadtree vacío con los límites y capacidad dados.
// capacity es el número máximo de puntos por nodo antes de subdividir.
func New(bounds Bounds, capacity int) *Quadtree {
	if capacity < 1 {
		capacity = 4
	}
	return &Quadtree{bounds: bounds, capacity: capacity}
}

// Insert agrega un punto al árbol. Retorna false si el punto está fuera de los límites.
func (qt *Quadtree) Insert(p Point) bool {
	if !qt.bounds.contains(p) {
		return false
	}

	if !qt.divided && len(qt.points) < qt.capacity {
		qt.points = append(qt.points, p)
		return true
	}

	if !qt.divided {
		qt.subdivide()
	}

	return qt.nw.Insert(p) || qt.ne.Insert(p) ||
		qt.sw.Insert(p) || qt.se.Insert(p)
}

// subdivide divide el nodo actual en 4 cuadrantes y redistribuye los puntos existentes.
func (qt *Quadtree) subdivide() {
	midX := (qt.bounds.MinX + qt.bounds.MaxX) / 2
	midY := (qt.bounds.MinY + qt.bounds.MaxY) / 2

	qt.nw = New(Bounds{qt.bounds.MinX, midY, midX, qt.bounds.MaxY}, qt.capacity)
	qt.ne = New(Bounds{midX, midY, qt.bounds.MaxX, qt.bounds.MaxY}, qt.capacity)
	qt.sw = New(Bounds{qt.bounds.MinX, qt.bounds.MinY, midX, midY}, qt.capacity)
	qt.se = New(Bounds{midX, qt.bounds.MinY, qt.bounds.MaxX, midY}, qt.capacity)
	qt.divided = true

	for _, p := range qt.points {
		qt.insertIntoChild(p)
	}
	qt.points = nil
}

// insertIntoChild intenta insertar p en cada hijo en orden NW→NE→SW→SE,
// parando en el primero que lo acepta (garantiza exactamente una inserción).
func (qt *Quadtree) insertIntoChild(p Point) {
	if qt.nw.Insert(p) {
		return
	}
	if qt.ne.Insert(p) {
		return
	}
	if qt.sw.Insert(p) {
		return
	}
	qt.se.Insert(p)
}

// QueryNearest retorna el punto más cercano a (x, y).
// Retorna nil si el árbol está vacío.
func (qt *Quadtree) QueryNearest(x, y float64) *Point {
	state := &nearestState{bestDist: math.MaxFloat64}
	qt.searchNearest(x, y, state)
	return state.best
}

type nearestState struct {
	best     *Point
	bestDist float64
}

// searchNearest recorre el árbol con poda: descarta ramas cuya distancia mínima
// al borde ya supera la mejor distancia encontrada hasta el momento.
func (qt *Quadtree) searchNearest(x, y float64, state *nearestState) {
	if qt.bounds.minDistToPoint(x, y) >= state.bestDist {
		return
	}

	for i := range qt.points {
		d := euclidean(qt.points[i].X, qt.points[i].Y, x, y)
		if d < state.bestDist {
			state.bestDist = d
			p := qt.points[i]
			state.best = &p
		}
	}

	if !qt.divided {
		return
	}

	// Visitar primero el cuadrante que contiene (x,y) para encontrar
	// una buena cota inicial rápido y podar más agresivamente.
	children := []*Quadtree{qt.nw, qt.ne, qt.sw, qt.se}
	for _, child := range children {
		if child.bounds.contains(Point{X: x, Y: y}) {
			child.searchNearest(x, y, state)
		}
	}
	for _, child := range children {
		if !child.bounds.contains(Point{X: x, Y: y}) {
			child.searchNearest(x, y, state)
		}
	}
}

// QueryRange retorna todos los puntos dentro del rectángulo dado.
func (qt *Quadtree) QueryRange(bounds Bounds) []Point {
	var results []Point
	qt.collectRange(bounds, &results)
	return results
}

func (qt *Quadtree) collectRange(bounds Bounds, results *[]Point) {
	// Poda: si los bounds no se solapan con este nodo, saltar
	if bounds.MaxX < qt.bounds.MinX || bounds.MinX > qt.bounds.MaxX ||
		bounds.MaxY < qt.bounds.MinY || bounds.MinY > qt.bounds.MaxY {
		return
	}

	for _, p := range qt.points {
		if bounds.contains(p) {
			*results = append(*results, p)
		}
	}

	if qt.divided {
		qt.nw.collectRange(bounds, results)
		qt.ne.collectRange(bounds, results)
		qt.sw.collectRange(bounds, results)
		qt.se.collectRange(bounds, results)
	}
}

// Size retorna el total de puntos almacenados en el árbol.
func (qt *Quadtree) Size() int {
	count := len(qt.points)
	if qt.divided {
		count += qt.nw.Size() + qt.ne.Size() + qt.sw.Size() + qt.se.Size()
	}
	return count
}

func euclidean(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}
