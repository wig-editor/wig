package mcwig

// Thanks: https://gist.github.com/pje/90e727f80685c78a6c1cfff35f62155a

type Line []rune

func (l Line) IsEmpty() bool {
	return len(l) == 0
}

func (l Line) Range(from, to int) []rune {
	if l.IsEmpty() {
		return []rune{}
	}
	if from < 0 {
		return []rune{}
	}
	if to > len(l) {
		return []rune{}
	}

	return l[from:to]
}

// the zero value is ready to use.
type List[T any] struct {
	root Element[T]
	Len  int
}

type Element[T any] struct {
	prev  *Element[T]
	next  *Element[T]
	list  *List[T]
	Value T
}

func (e *Element[T]) Next() *Element[T] {
	n := e.next
	if e.list == nil || n == &e.list.root {
		return nil
	}
	return n
}

func (e *Element[T]) Prev() *Element[T] {
	p := e.prev
	if e.list == nil || p == &e.list.root {
		return nil
	}
	return p
}

func (l *List[T]) First() *Element[T] {
	if l.Len == 0 {
		return nil
	}
	return l.root.next
}

func (l *List[T]) Last() *Element[T] {
	if l.Len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *List[T]) PushFront(v T) *Element[T] {
	if l.root.next == nil {
		l.init()
	}
	return l.insertValueAfter(v, &l.root)
}

func (l *List[T]) PushBack(v T) *Element[T] {
	if l.root.next == nil {
		l.init()
	}
	return l.insertValueAfter(v, l.root.prev)
}

func (l *List[T]) Remove(e *Element[T]) T {
	if e.list == l {
		l.remove(e)
	}
	return e.Value
}

// Constructs a new List[T] from the given slice, returns the List[T].
func FromSlice[T any](slice []T) List[T] {
	var list List[T]
	for _, e := range slice {
		list.PushFront(e)
	}
	return list
}

func (l *List[T]) init() {
	l.root = *new(Element[T])
	l.root.next = &l.root
	l.root.prev = &l.root
}

func (l *List[T]) insertAfter(e *Element[T], at *Element[T]) *Element[T] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.Len++
	return e
}

func (l *List[T]) insertValueAfter(v T, at *Element[T]) *Element[T] {
	e := Element[T]{Value: v}
	return l.insertAfter(&e, at)
}

func (l *List[T]) remove(e *Element[T]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil
	l.Len--
}
