package hashgrid

import (
	"math"
	"slices"
)

const offset = 1 << 31

// GridEntry defines the minimum interface required for items stored in the grid.
type GridEntry interface {
	comparable

	Min() (x, y float32)
	Max() (x, y float32)
}

// GridKey is a compact representation of grid cell coordinates.
type GridKey uint64

func EncodeGridKey(x, y int32) GridKey {
	ux := uint32(int64(x) + offset)
	uy := uint32(int64(y) + offset)
	return GridKey(uint64(ux)<<32 | uint64(uy))
}

func DecodeGridKey(key GridKey) (x, y int32) {
	x = int32(int64(uint32(key>>32)) - offset)
	y = int32(int64(uint32(key)) - offset)
	return x, y
}

// Grid is a spatial hash grid for efficient spatial queries.
type Grid[T GridEntry] struct {
	cellSize     float32
	cells        map[GridKey][]T
	cellKeys     map[T][]GridKey
	queryResults []T
	queryKeys    []GridKey
	getKeys      []GridKey
}

type CellCheckFunc func(minX, minY, maxX, maxY float32) bool

// New creates a new Grid with the specified cell size.
//
// Panics if cellSize is zero or negative.
func New[T GridEntry](cellSize float32) *Grid[T] {
	if cellSize <= 0 {
		panic("cellSize must be positive")
	}
	return &Grid[T]{
		cellSize:     cellSize,
		cells:        make(map[GridKey][]T),
		cellKeys:     make(map[T][]GridKey),
		queryResults: make([]T, 0, 32),
		queryKeys:    make([]GridKey, 0, 16),
		getKeys:      make([]GridKey, 0, 16),
	}
}

// CellSize returns the size of each grid cell.
func (g *Grid[T]) CellSize() float32 {
	return g.cellSize
}

// Keys returns a slice of all occupied grid keys.
// Returns a copy of the keys that the caller owns.
func (g *Grid[T]) Keys(minX, minY, maxX, maxY float32) []GridKey {
	keys := g.KeysUnsafe(minX, minY, maxX, maxY)
	return slices.Clone(keys)
}

// KeysUnsafe returns a slice of all occupied grid keys.
// The returned slice is only valid until the next call to KeysUnsafe on this grid.
// This method provides zero-allocation key queries for performance-critical code.
func (g *Grid[T]) KeysUnsafe(minX, minY, maxX, maxY float32) []GridKey {
	g.queryKeys = g.queryKeys[:0]

	sX, sY, eX, eY := computeCellRange(minX, minY, maxX, maxY, g.cellSize)
	for x := sX; x < eX; x++ {
		for y := sY; y < eY; y++ {
			key := EncodeGridKey(x, y)

			if _, ok := g.cells[key]; ok {
				g.queryKeys = append(g.queryKeys, key)
			}
		}
	}

	return g.queryKeys
}

// Insert adds an entry to all grid cells it occupies.
func (g *Grid[T]) Insert(entry T) bool {
	return g.InsertStrictCheck(entry, nil)
}

func (g *Grid[T]) InsertStrictCheck(entry T, check CellCheckFunc) bool {
	if _, exists := g.cellKeys[entry]; exists {
		return false
	}

	keys := g.GetKeys(entry, check)
	g.cellKeys[entry] = keys

	for i := range keys {
		g.cells[keys[i]] = append(g.cells[keys[i]], entry)
	}

	return true
}

// Remove removes an entry from all grid cells it occupies.
func (g *Grid[T]) Remove(entry T) {
	if keys, exists := g.cellKeys[entry]; exists {
		for i := range keys {
			entities := g.cells[keys[i]]
			for j, e := range entities {
				if e == entry {
					g.cells[keys[i]] = append(entities[:j], entities[j+1:]...)
					if len(g.cells[keys[i]]) == 0 {
						delete(g.cells, keys[i])
					}
					break
				}
			}
		}
		delete(g.cellKeys, entry)
	}
}

// Query returns all entries that intersect the specified rectangular region.
// Returns a copy of the results that the caller owns.
func (g *Grid[T]) Query(minX, minY, maxX, maxY float32) []T {
	results := g.QueryUnsafe(minX, minY, maxX, maxY)
	return slices.Clone(results)
}

// QueryUnsafe returns all entries that intersect the specified rectangular region.
// The returned slice is only valid until the next call to QueryUnsafe on this grid.
// This method provides zero-allocation queries for performance-critical code.
func (g *Grid[T]) QueryUnsafe(minX, minY, maxX, maxY float32) []T {
	g.queryResults = g.queryResults[:0]

	sX, sY, eX, eY := computeCellRange(minX, minY, maxX, maxY, g.cellSize)
	for x := sX; x < eX; x++ {
		for y := sY; y < eY; y++ {
			key := EncodeGridKey(x, y)

			if entries, exists := g.cells[key]; exists {
				for i := range entries {
					if !slices.Contains(g.queryResults, entries[i]) {
						g.queryResults = append(g.queryResults, entries[i])
					}
				}
			}
		}
	}

	return g.queryResults
}

// Clear removes all entries from the grid.
func (g *Grid[T]) Clear() {
	g.cells = make(map[GridKey][]T)
	g.cellKeys = make(map[T][]GridKey)
	g.queryResults = g.queryResults[:0]
	g.queryKeys = g.queryKeys[:0]
	g.getKeys = g.getKeys[:0]
}

// GetKeys computes all grid keys that the entry occupies.
// Returns a copy that the caller owns.
func (g *Grid[T]) GetKeys(entry T, check CellCheckFunc) []GridKey {
	keys := g.getKeysInternal(entry, check)

	result := make([]GridKey, len(keys))
	copy(result, keys)

	return result
}

func (g *Grid[T]) getKeysInternal(entry T, check CellCheckFunc) []GridKey {
	minX, minY := entry.Min()
	maxX, maxY := entry.Max()

	sX, sY, eX, eY := computeCellRange(minX, minY, maxX, maxY, g.cellSize)

	g.getKeys = g.getKeys[:0]

	for x := sX; x < eX; x++ {
		for y := sY; y < eY; y++ {
			if check != nil {
				if !check(float32(x)*g.cellSize, float32(y)*g.cellSize, float32(x+1)*g.cellSize, float32(y+1)*g.cellSize) {
					continue
				}
			}
			key := EncodeGridKey(x, y)
			g.getKeys = append(g.getKeys, key)
		}
	}

	return g.getKeys
}

func computeCellRange(minX, minY, maxX, maxY, cellSize float32) (startX, startY, endX, endY int32) {
	startX = int32(math.Floor(float64(minX / cellSize)))
	startY = int32(math.Floor(float64(minY / cellSize)))
	endX = int32(math.Ceil(float64(maxX / cellSize)))
	endY = int32(math.Ceil(float64(maxY / cellSize)))
	return
}
