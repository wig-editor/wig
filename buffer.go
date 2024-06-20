package mcwig

import (
	"bytes"
	"os"
	"strings"
)

type Mode int

const (
	MODE_NORMAL      Mode = 0
	MODE_INSERT      Mode = 1
	MODE_VISUAL      Mode = 2
	MODE_VISUAL_LINE Mode = 3
)

func (m Mode) String() string {
	if m == MODE_NORMAL {
		return "NOR"
	}
	if m == MODE_INSERT {
		return "INS"
	}
	if m == MODE_VISUAL_LINE {
		return "VIS LINE"
	}
	return "VIS"
}

// Driver represents anything that can run selected text. it can be sql conncetion,
// or rest client.
type Driver interface {
	// Execute thing under cursor: line or seleciton
	Exec(*Editor, *Buffer, *Element[Line])
	// Execute whole buffer
	ExecBuffer()
}

type Buffer struct {
	Mode         Mode
	FilePath     string
	ScrollOffset int
	Lines        List[Line]
	Cursor       Cursor
	Selection    *Selection
	Driver       Driver
	IndentCh     []rune
}

func NewBuffer() *Buffer {
	lines := List[Line]{}
	lines.PushBack(Line{})
	return &Buffer{
		Lines:     lines,
		Cursor:    Cursor{0, 0, 0},
		IndentCh:  []rune{'\t'},
		Selection: nil,
		Driver:    nil,
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
	buf.Lines = List[Line]{}

	for _, line := range bytes.Split(data, []byte("\n")) {
		buf.Lines.PushBack([]rune(string(line)))
	}

	return buf, nil
}

func (b *Buffer) GetName() string {
	if len(b.FilePath) > 0 {
		return b.FilePath
	}
	return "[No Name]"
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
func (e *Editor) BufferFindByFilePath(fp string) *Buffer {
	for _, b := range e.Buffers {
		if b.FilePath == fp {
			return b
		}
	}

	b := NewBuffer()
	b.FilePath = fp
	b.Lines = List[Line]{}
	e.Buffers = append(e.Buffers, b)
	return b
}

func (b *Buffer) Append(s string) {
	for _, line := range strings.Split(s, "\n") {
		b.Lines.PushBack([]rune(line))
	}
}

// Remove all lines
func (b *Buffer) ResetLines() {
	l := b.Lines.First()
	for l != nil {
		next := l.Next()
		l.Value = nil
		b.Lines.Remove(l)
		l = next
	}
}

func (b *Buffer) String() string {
	buf := bytes.NewBuffer(nil)

	line := b.Lines.First()
	sep := "\n"
	for line != nil {
		if line.Next() == nil {
			sep = ""
		}
		buf.WriteString(string(line.Value) + sep)
		line = line.Next()
	}

	return buf.String()
}
