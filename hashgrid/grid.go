package hashgrid

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
	left, right, top, bottom uint8
	cellSize                 float32
	cells                    map[GridKey][]T
}

type GridPadding uint8

const (
	NoPadding   GridPadding = 0
	LeftPadding GridPadding = 1 << (iota - 1)
	RightPadding
	TopPadding
	BottomPadding
	AllPadding = LeftPadding | RightPadding | TopPadding | BottomPadding
)

// New creates a new Grid with the specified cell size.
//
// Panics if cellSize is zero or negative.
func New[T GridEntry](cellSize float32) *Grid[T] {
	return NewWithPadding[T](cellSize, NoPadding)
}

// NewWithPadding creates a new Grid with the specified cell size and padding.
//
// Panics if cellSize is zero or negative.
func NewWithPadding[T GridEntry](cellSize float32, padding GridPadding) *Grid[T] {
	if cellSize <= 0 {
		panic("cellSize must be positive")
	}
	return &Grid[T]{
		left:     uint8(padding & LeftPadding),
		right:    uint8((padding & RightPadding) >> 1),
		top:      uint8((padding & TopPadding) >> 2),
		bottom:   uint8((padding & BottomPadding) >> 3),
		cellSize: cellSize,
		cells:    make(map[GridKey][]T),
	}
}

// CellSize returns the size of each grid cell.
func (g *Grid[T]) CellSize() float32 {
	return g.cellSize
}

// Keys returns a slice of all occupied grid keys.
func (g *Grid[T]) Keys() []GridKey {
	keys := make([]GridKey, 0, len(g.cells))
	for k := range g.cells {
		keys = append(keys, k)
	}
	return keys
}

// Insert adds an entry to all grid cells it occupies.
func (g *Grid[T]) Insert(entry T) {
	keys := g.GetKeys(entry)
	for _, key := range keys {
		entries := g.cells[key]
		// Avoid duplicates
		duplicate := false
		for i := range entries {
			if entries[i] == entry {
				duplicate = true
				break
			}
		}
		if !duplicate {
			g.cells[key] = append(entries, entry)
		}
	}
}

// Remove removes an entry from all grid cells it occupies.
func (g *Grid[T]) Remove(entry T) {
	keys := g.GetKeys(entry)
	for _, key := range keys {
		entries := g.cells[key]
		for j, e := range entries {
			if e == entry {
				g.cells[key] = append(entries[:j], entries[j+1:]...)
				if len(g.cells[key]) == 0 {
					delete(g.cells, key)
				}
				break
			}
		}
	}
}

// Query returns all entries that intersect the specified rectangular region.
func (g *Grid[T]) Query(minX, minY, maxX, maxY float32) []T {
	startX := int32(minX / g.cellSize)
	startY := int32(minY / g.cellSize)
	endX := int32(maxX / g.cellSize)
	endY := int32(maxY / g.cellSize)

	seen := make(map[T]struct{})
	var results []T

	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			for _, entry := range g.cells[EncodeGridKey(x, y)] {
				if _, exists := seen[entry]; exists {
					continue
				}
				eMinX, eMinY := entry.Min()
				eMaxX, eMaxY := entry.Max()
				if eMaxX >= minX && eMinX <= maxX && eMaxY >= minY && eMinY <= maxY {
					seen[entry] = struct{}{}
					results = append(results, entry)
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

	startX := int32(minX/g.cellSize) - int32(g.left)
	startY := int32(minY/g.cellSize) - int32(g.top)
	endX := (int32(maxX/g.cellSize) - 1) + int32(g.right)
	endY := (int32(maxY/g.cellSize) - 1) + int32(g.bottom)

	var keys []GridKey
	for x := startX; x <= endX; x++ {
		for y := startY; y <= endY; y++ {
			keys = append(keys, EncodeGridKey(x, y))
		}
	}
	return keys
}
