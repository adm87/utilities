package hash

import "math"

func EncodeGridKey(x, y int32) uint64 {
	const offset = 1 << 31
	upper := uint64(uint32(int64(x) + offset))
	lower := uint64(uint32(int64(y) + offset))
	return (upper << 32) | lower
}

// Grid is a simple spatial hash grid that stores items in cells based on their coordinates.
//
// Use this for static or infrequently-updated items. It is a minimalist implementation.
type Grid[T comparable] struct {
	gen        uint64
	cellWidth  float32
	cellHeight float32
	cells      map[uint64][]T
	items      map[T]uint64
	itemCells  map[T][]uint64
	qBuf       []T
}

func NewGrid[T comparable](cellWidth, cellHeight float32) *Grid[T] {
	return &Grid[T]{
		cellWidth:  cellWidth,
		cellHeight: cellHeight,
		cells:      make(map[uint64][]T),
		items:      make(map[T]uint64),
		itemCells:  make(map[T][]uint64),
	}
}

func (g *Grid[T]) cellRange(minX, minY, maxX, maxY float32) (minCellX, minCellY, maxCellX, maxCellY int32) {
	minCellX = int32(math.Floor(float64(minX / g.cellWidth)))
	minCellY = int32(math.Floor(float64(minY / g.cellHeight)))
	maxCellX = int32(math.Ceil(float64(maxX / g.cellWidth)))
	maxCellY = int32(math.Ceil(float64(maxY / g.cellHeight)))
	return
}

// ForEach calls the given function for each item in the grid.
func (g *Grid[T]) ForEach(fn func(item T)) {
	for item := range g.items {
		fn(item)
	}
}

// Clear removes all items from the grid.
func (g *Grid[T]) Clear() {
	clear(g.cells)
	clear(g.items)
	clear(g.itemCells)
	clear(g.qBuf)
	g.gen = 0
}

// Resize changes the cell size of the grid, clearing all existing items.
// If the size is unchanged, no action is taken.
//
// WARNING: This will remove all existing items in the grid.
func (g *Grid[T]) Resize(cellWidth, cellHeight float32) {
	if g.cellWidth == cellWidth && g.cellHeight == cellHeight {
		return
	}
	g.Clear()

	g.cellWidth = cellWidth
	g.cellHeight = cellHeight
}

// Contains checks if the item is already in the grid.
func (g *Grid[T]) Contains(item T) bool {
	_, exists := g.items[item]
	return exists
}

// Insert adds an item to the grid. Returns false if the item was already present.
func (g *Grid[T]) Insert(item T, minX, minY, maxX, maxY float32) bool {
	if g.Contains(item) {
		return false
	}

	var cellKeys []uint64
	minCellX, minCellY, maxCellX, maxCellY := g.cellRange(minX, minY, maxX, maxY)
	for cy := minCellY; cy < maxCellY; cy++ {
		for cx := minCellX; cx < maxCellX; cx++ {
			key := EncodeGridKey(cx, cy)
			g.cells[key] = append(g.cells[key], item)
			cellKeys = append(cellKeys, key)
		}
	}

	g.items[item] = 0
	g.itemCells[item] = cellKeys

	return true
}

// Remove removes an item from the grid.
//
// Cells are removed if they are no longer storing an item.
func (g *Grid[T]) Remove(item T) {
	if !g.Contains(item) {
		return
	}

	cellKeys := g.itemCells[item]
	delete(g.items, item)
	delete(g.itemCells, item)

	for _, key := range cellKeys {
		items := g.cells[key]

		// compact in-place, keeping only elements != item
		j := 0
		for _, it := range items {
			if it != item {
				items[j] = it
				j++
			}
		}

		if j == 0 {
			delete(g.cells, key)
		} else {
			g.cells[key] = items[:j]
		}
	}
}

// Query returns all items that intersect the given AABB.
func (g *Grid[T]) Query(minX, minY, maxX, maxY float32) []T {
	g.qBuf = g.qBuf[:0]
	g.gen++

	minCellX, minCellY, maxCellX, maxCellY := g.cellRange(minX, minY, maxX, maxY)
	for cy := minCellY; cy < maxCellY; cy++ {
		for cx := minCellX; cx < maxCellX; cx++ {
			key := EncodeGridKey(cx, cy)
			items, exists := g.cells[key]
			if !exists {
				continue
			}

			for _, item := range items {
				if g.items[item] != g.gen {
					g.qBuf = append(g.qBuf, item)
					g.items[item] = g.gen
				}
			}
		}
	}

	return g.qBuf
}
