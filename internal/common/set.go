package common

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	s := make(Set[T])
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) Remove(v T) bool {
	_, ok := s[v]
	if ok {
		delete(s, v)
	}
	return ok
}

func (s Set[T]) Clear(v T) {
	clear(s)
}
