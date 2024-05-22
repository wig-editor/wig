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
}

func NewBuffer() *Buffer {
	return &Buffer{
		Lines: List[Line]{},
	}
}

func BufferReadFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	buf := NewBuffer()
	buf.FilePath = path
	buf.Cursor = Cursor{0, 0, 0}
	buf.Selection = nil

	for _, line := range bytes.Split(data, []byte("\n")) {
		buf.Lines.PushBack([]rune(string(line)))
	}

	return buf, nil
}
