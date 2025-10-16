package hashgrid

// GridEntry defines the minimum interface required for items stored in the grid.
type GridEntry interface {
	comparable

	Min() (x, y float32)
	Max() (x, y float32)
}

// GridKey is a compact representation of grid cell coordinates.
type GridKey uint64

func NewGridKey(x, y int) GridKey {
	return GridKey(uint64(uint32(x))<<32 | uint64(uint32(y)))
}

// Grid is a spatial hash grid for efficient spatial queries.
type Grid[T GridEntry] struct {
	cellSize float32
	cells    map[GridKey][]T
}

// New creates a new Grid with the specified cell size.
//
// Panics if cellSize is zero or negative.
func New[T GridEntry](cellSize float32) *Grid[T] {
	if cellSize <= 0 {
		panic("cellSize must be positive")
	}
	return &Grid[T]{
		cellSize: cellSize,
		cells:    make(map[GridKey][]T),
	}
}

// Insert adds an entry to all grid cells it occupies.
func (g *Grid[T]) Insert(entry T) {
	keys := g.GetKeys(entry)
	for i := range keys {
		key := keys[i]
		if g.cells[key] == nil {
			g.cells[key] = make([]T, 0)
		}

		for j := range g.cells[key] {
			if g.cells[key][j] == entry {
				goto nextKey
			}
		}

		g.cells[key] = append(g.cells[key], entry)
	nextKey:
	}
}

// Query returns all entries that intersect the specified rectangular region.
func (g *Grid[T]) Remove(entry T) {
	keys := g.GetKeys(entry)
	for i := range keys {
		key := keys[i]
		for j := range g.cells[key] {
			if g.cells[key][j] == entry {
				g.cells[key] = append(g.cells[key][:j], g.cells[key][j+1:]...)
				break
			}
		}
		if len(g.cells[key]) == 0 {
			delete(g.cells, key)
		}
	}
}

// Query returns all entries that intersect the specified rectangular region.
func (g *Grid[T]) Query(minX, minY, maxX, maxY float32) []T {
	startX := int(minX / g.cellSize)
	startY := int(minY / g.cellSize)
	endX := int(maxX / g.cellSize)
	endY := int(maxY / g.cellSize)

	seen := make(map[T]struct{})

	var results []T
	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			key := NewGridKey(x, y)
			for e := range g.cells[key] {
				if _, exists := seen[g.cells[key][e]]; exists {
					continue // Already processed
				}

				eMinX, eMinY := g.cells[key][e].Min()
				eMaxX, eMaxY := g.cells[key][e].Max()
				if eMaxX >= minX && eMinX <= maxX && eMaxY >= minY && eMinY <= maxY {
					seen[g.cells[key][e]] = struct{}{}
					results = append(results, g.cells[key][e])
				}
			}
		}
	}
	return results
}

// Clear removes all entries from the grid.
func (g *Grid[T]) Clear() {
	g.cells = make(map[GridKey][]T)
}

// GetKeys computes all grid keys that the entry occupies.
func (g *Grid[T]) GetKeys(entry T) []GridKey {
	minX, minY := entry.Min()
	maxX, maxY := entry.Max()

	startX := int(minX / g.cellSize)
	startY := int(minY / g.cellSize)
	endX := int(maxX / g.cellSize)
	endY := int(maxY / g.cellSize)

	var keys []GridKey
	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			keys = append(keys, NewGridKey(x, y))
		}
	}
	return keys
}
