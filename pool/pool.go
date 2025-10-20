package pool

// Pool is a generic object pool for reusing objects of Poolable types.
//
// Use sync.Pool from the standard library for concurrent use cases.
type Pool[T any] struct {
	items []T

	New func() T
}

func (p *Pool[T]) Get() T {
	n := len(p.items)
	if n == 0 {
		return p.New()
	}
	item := p.items[n-1]
	p.items = p.items[:n-1]
	return item
}

func (p *Pool[T]) Put(item T) {
	p.items = append(p.items, item)
}

func (p *Pool[T]) Len() int {
	return len(p.items)
}
