package urlutils

import (
	"strings"
)

const pathVariable = ":"

type PathTree[T any] struct {
	root *pathSegment[T]
}

func NewPathTree[T any]() *PathTree[T] { return &PathTree[T]{root: newPathSegment[T]()} }

func (t *PathTree[T]) Add(path string, value T)   { t.root.add(path, value) }
func (t *PathTree[T]) Find(path string) (T, bool) { return t.root.find(path) }

func (t *PathTree[T]) LongestCommonPrefix(path string) (T, bool) {
	return t.root.longestCommonPrefix(path)
}

type pathSegment[T any] struct {
	children map[string]*pathSegment[T]
	isEnd    bool

	// пустое, если isEnd == false
	value T
}

func newPathSegment[T any]() *pathSegment[T] {
	var empty T
	return &pathSegment[T]{
		isEnd:    false,
		value:    empty,
		children: make(map[string]*pathSegment[T]),
	}
}

func isPathVariable(seg string) bool {
	return strings.HasPrefix(seg, ":")
}

func (s *pathSegment[T]) hasPathVariable() bool {
	_, ok := s.children[pathVariable]
	return ok
}

func (s *pathSegment[T]) add(path string, value T) {
	segments := strings.Split(path, "/")
	if len(segments) < 1 {
		return
	}

	cur := s
	for _, seg := range segments[1:] {
		if isPathVariable(seg) {
			seg = pathVariable
		}

		child, ok := cur.children[seg]
		if !ok {
			cur.children[seg] = newPathSegment[T]()
			child = cur.children[seg]
		}
		cur = child
	}
	cur.value, cur.isEnd = value, true
}

func (s *pathSegment[T]) find(path string) (T, bool) {
	val, isEnd := s.value, s.isEnd

	segments := strings.Split(path, "/")
	if len(segments) < 1 {
		return val, isEnd
	}

	cur := s
	for _, seg := range segments[1:] {
		child, ok := cur.children[seg]
		if !ok && cur.hasPathVariable() {
			child = cur.children[pathVariable]
		} else if !ok {
			break
		}
		cur = child
	}
	return val, isEnd
}

func (s *pathSegment[T]) longestCommonPrefix(path string) (T, bool) {
	segments := strings.Split(path, "/")
	if len(segments) < 1 {
		return s.value, s.isEnd
	}

	cur, curVal, isEnd := s, s.value, s.isEnd
	for _, seg := range segments[1:] {
		child, ok := cur.children[seg]
		if !ok && cur.hasPathVariable() {
			child = cur.children[pathVariable]
		} else if !ok {
			break
		}

		cur = child
		if cur.isEnd {
			curVal, isEnd = cur.value, cur.isEnd
		}
	}
	return curVal, isEnd
}
