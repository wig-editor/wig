package mcwig

type Stack[T any] struct {
	elements []T
}

func (s *Stack[T]) Push(element T) {
	s.elements = append(s.elements, element)
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zeroValue T
		return zeroValue, false
	} else {
		index := len(s.elements) - 1
		element := s.elements[index]
		s.elements = s.elements[:index]
		return element, true
	}
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}
