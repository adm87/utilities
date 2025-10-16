package linq

// ========== Batch ==========

func Batch[T any](items []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	var batches [][]T
	for i := 0; i < len(items); i += size {
		end := min(i+size, len(items))
		batches = append(batches, items[i:end])
	}

	return batches
}

// ========== Distinct ==========

func Distinct[T comparable](items []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(items))

	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
