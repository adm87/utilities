package hash

// Grid is a simple spatial hash grid that stores items in cells based on their coordinates.
//
// Use this for static or infrequently-updated items. It is a minimalist implementation.
type Grid[T comparable] struct {
	gen       uint64
	cellSize  float32
	cells     map[uint64][]T
	items     map[T]uint64
	itemCells map[T][]uint64
	qBuf      []T
}

func NewGrid[T comparable](cellSize float32) *Grid[T] {
	return &Grid[T]{
		cellSize:  cellSize,
		cells:     make(map[uint64][]T),
		items:     make(map[T]uint64),
		itemCells: make(map[T][]uint64),
	}
}

func (g *Grid[T]) cellKey(x, y int32) uint64 {
	const offset = 1 << 31
	upper := uint64(uint32(int64(x) + offset))
	lower := uint64(uint32(int64(y) + offset))
	return (upper << 32) | lower
}

func (g *Grid[T]) cellRange(minX, minY, maxX, maxY float32) (minCellX, minCellY, maxCellX, maxCellY int32) {
	minCellX = int32(minX / g.cellSize)
	minCellY = int32(minY / g.cellSize)
	maxCellX = int32(maxX / g.cellSize)
	maxCellY = int32(maxY / g.cellSize)
	return
}

func (g *Grid[T]) Contains(item T) bool {
	_, exists := g.items[item]
	return exists
}

func (g *Grid[T]) Insert(item T, minX, minY, maxX, maxY float32) bool {
	if g.Contains(item) {
		return false
	}

	var cellKeys []uint64
	minCellX, minCellY, maxCellX, maxCellY := g.cellRange(minX, minY, maxX, maxY)
	for cy := minCellY; cy <= maxCellY; cy++ {
		for cx := minCellX; cx <= maxCellX; cx++ {
			key := g.cellKey(cx, cy)
			g.cells[key] = append(g.cells[key], item)
			cellKeys = append(cellKeys, key)
		}
	}

	g.items[item] = 0
	g.itemCells[item] = cellKeys

	return true
}

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

func (g *Grid[T]) Query(minX, minY, maxX, maxY float32) []T {
	g.qBuf = g.qBuf[:0]
	g.gen++

	minCellX, minCellY, maxCellX, maxCellY := g.cellRange(minX, minY, maxX, maxY)
	for cy := minCellY; cy <= maxCellY; cy++ {
		for cx := minCellX; cx <= maxCellX; cx++ {
			key := g.cellKey(cx, cy)
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
