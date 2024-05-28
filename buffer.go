package mcwig

import (
	"bytes"
	"os"
)

type Mode int

const (
	MODE_NORMAL Mode = 0
	MODE_INSERT Mode = 1
	MODE_VISUAL Mode = 2
)

func (m Mode) String() string {
	if m == MODE_NORMAL {
		return "NOR"
	}
	if m == MODE_INSERT {
		return "INS"
	}
	return "VIS"
}

type Cursor struct {
	Line                 int
	Char                 int
	PreserveCharPosition int
}

type Buffer struct {
	Mode         Mode
	FilePath     string
	ScrollOffset int
	Lines        List[Line]
	Cursor       Cursor
	Selection    *Selection

	Name string
}

func NewBuffer() *Buffer {
	return &Buffer{
		Lines:     List[Line]{},
		Cursor:    Cursor{0, 0, 0},
		Selection: nil,
	}
}

func BufferReadFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	buf := NewBuffer()
	buf.FilePath = path
	buf.Name = path
	buf.Cursor = Cursor{0, 0, 0}
	buf.Selection = nil

	for _, line := range bytes.Split(data, []byte("\n")) {
		buf.Lines.PushBack([]rune(string(line)))
	}

	return buf, nil
}

func (b *Buffer) GetName() string {
	if len(b.Name) > 0 {
		return b.Name
	}
	return b.FilePath
}

func (b *Buffer) AppendStringLine(s string) {
	b.Lines.PushBack([]rune(string(s)))
}

func (b *Buffer) Save() error {
	f, err := os.Create(b.FilePath)
	if err != nil {
		return err
	}

	line := b.Lines.First()
	sep := "\n"
	for line != nil {
		if line.Next() == nil {
			sep = ""
		}
		f.WriteString(string(line.Value) + sep)
		line = line.Next()
	}

	return nil
}

// Find or create new buffer
func (e *Editor) BufferGetByName(name string) *Buffer {
	for _, b := range e.Buffers {
		if b.Name == name {
			return b
		}
	}

	b := NewBuffer()
	b.Name = name
	e.Buffers = append(e.Buffers, b)
	return b
}
