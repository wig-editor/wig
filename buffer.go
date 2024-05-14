package mcwig

import (
	"bytes"
	"os"
)

type Mode int

const MODE_NORMAL Mode = 0
const MODE_INSERT Mode = 1

func (m Mode) String() string {
	if m == MODE_NORMAL {
		return "NOR"
	}
	return "INS"
}

type Line struct {
	Data []rune
	Next *Line
	Prev *Line
}

func (l *Line) String() string {
	return string(l.Data)
}

type LineList struct {
	Head *Line
	Tail *Line
	Size int
}

func (ll *LineList) Append(data []rune) *Line {
	line := &Line{Data: data}
	if ll.Head == nil {
		ll.Head = line
		ll.Tail = line
	} else {
		ll.Tail.Next = line
		line.Prev = ll.Tail
		ll.Tail = line
	}
	ll.Size++
	return line
}

func (ll *LineList) Insert(data []rune, index int) *Line {
	if index < 0 || index > ll.Size {
		return nil
	}

	line := &Line{Data: data}
	if ll.Size == 0 {
		ll.Head = line
		ll.Tail = line
	} else if index == 0 {
		line.Next = ll.Head
		ll.Head.Prev = line
		ll.Head = line
	} else if index == ll.Size {
		ll.Tail.Next = line
		line.Prev = ll.Tail
		ll.Tail = line
	} else {
		current := ll.Head
		for i := 0; i < index-1; i++ {
			current = current.Next
		}
		line.Prev = current
		line.Next = current.Next
		current.Next.Prev = line
		current.Next = line
	}
	ll.Size++
	return line
}

func (ll *LineList) Delete(index int) {
	if index < 0 || index >= ll.Size {
		return
	}

	if ll.Size == 1 {
		ll.Head = nil
		ll.Tail = nil
	} else if index == 0 {
		ll.Head = ll.Head.Next
		ll.Head.Prev = nil
	} else if index == ll.Size-1 {
		ll.Tail = ll.Tail.Prev
		ll.Tail.Next = nil
	} else {
		current := ll.Head
		for i := 0; i < index; i++ {
			current = current.Next
		}
		current.Prev.Next = current.Next
		current.Next.Prev = current.Prev
	}
	ll.Size--
}

func (ll *LineList) String() string {
	var buf bytes.Buffer
	current := ll.Head
	for current != nil {
		buf.WriteString(string(current.Data))
		buf.WriteString("\n")
		current = current.Next
	}
	return buf.String()
}

type Cursor struct {
	Line                 int
	Char                 int
	PreserveCharPosition int
}

type Buffer struct {
	FilePath     string
	Mode         Mode
	ScrollOffset int
	Lines        *LineList
	Cursor       Cursor
}

func NewBuffer() *Buffer {
	return &Buffer{
		Lines: &LineList{},
	}
}

func BufferReadFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := &LineList{}
	buf := &Buffer{
		FilePath: path,
		Lines:    lines,
		Cursor:   Cursor{3, 3, 3},
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
		lines.Append([]rune(string(line)))
	}

	return buf, nil
}
