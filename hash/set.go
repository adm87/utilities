package hash

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](size ...int) Set[T] {
	sz := 0
	if len(size) > 0 {
		sz = size[0]
	}
	set := make(Set[T], sz)
	return set
}

func (s *Set[T]) Add(item T) bool {
	if s.Contains(item) {
		return false
	}
	(*s)[item] = struct{}{}
	return true
}

func (s *Set[T]) Remove(item T) {
	delete(*s, item)
}

func (s *Set[T]) Contains(item T) bool {
	_, exists := (*s)[item]
	return exists
}

func (s *Set[T]) Size() int {
	return len(*s)
}

func (s *Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(*s))
	for item := range *s {
		slice = append(slice, item)
	}
	return slice
}
