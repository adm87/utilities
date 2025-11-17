package hash

import (
	"math/rand"
	"testing"
)

type TestItem struct {
	ID int
}

func generateItems(count int) []TestItem {
	items := make([]TestItem, count)
	for i := range items {
		items[i] = TestItem{ID: i}
	}
	return items
}

func BenchmarkMinimalistGridInsert(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		item := items[i%len(items)]
		x := rand.Float32() * 3200
		y := rand.Float32() * 3200
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}
}

func BenchmarkMinimalistGridQuery(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(1000)

	// Pre-populate grid
	for _, item := range items {
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = grid.Query([4]float32{100, 100, 300, 300})
	}
}

func BenchmarkMinimalistGridQueryEmpty(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(100)

	// Pre-populate grid in small area (0-640 range)
	for _, item := range items {
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Query empty region far away
		_ = grid.Query([4]float32{5000, 5000, 5300, 5300})
	}
}

func BenchmarkMinimalistGridRemove(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(1000)

	// Pre-populate grid once
	for _, item := range items {
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		item := items[i%len(items)]
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		grid.Remove(item)

		// Re-insert to maintain state
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}
}

func BenchmarkMinimalistGridContains(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(1000)

	// Pre-populate grid
	for _, item := range items {
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		item := items[i%len(items)]
		_ = grid.Contains(item)
	}
}

func BenchmarkMinimalistGridLargeQuery(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(10000)

	// Pre-populate large grid
	for _, item := range items {
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		grid.Insert(item, [4]float32{x, y, x + 32, y + 32}, NoGridPadding)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Large query that spans many cells
		_ = grid.Query([4]float32{0, 0, 1000, 1000})
	}
}

func BenchmarkMinimalistGridSpanningInsert(b *testing.B) {
	grid := NewGrid[TestItem](64.0, 64.0)
	items := generateItems(1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		item := items[i%len(items)]
		x := rand.Float32() * 2048
		y := rand.Float32() * 2048
		// Large item spanning multiple cells
		grid.Insert(item, [4]float32{x, y, x + 200, y + 200}, NoGridPadding)
	}
}
