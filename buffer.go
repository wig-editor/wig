package wig

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

type Mode int

const (
	MODE_NORMAL       Mode = 0
	MODE_INSERT       Mode = 1
	MODE_VISUAL       Mode = 2
	MODE_VISUAL_LINE  Mode = 3
	MODE_VISUAL_BLOCK Mode = 4
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
	if m == MODE_VISUAL_BLOCK {
		return "VIS BLOCK"
	}
	return "VIS"
}

// Driver represents anything that can run selected text. it can be sql connection,
// or rest client.
type Driver interface {
	// Execute thing under cursor: line or seleciton
	Exec(*Editor, *Buffer, *Element[Line])
	// Execute whole buffer
	ExecBuffer()
}

type VisualBlockInsertState struct {
	StartLine int
	EndLine   int
	Char      int
}

type BlameInfo struct {
	Hash    string
	Author  string
	Display string
}

type Buffer struct {
	mode              Mode
	FilePath          string
	Lines             List[Line]
	Selection         *Selection
	Driver            Driver
	IndentCh          []rune
	Tx                *Transaction
	UndoRedo          *UndoRedo
	Highlighter       Highlighter
	KeyHandler        *KeyHandler
	GitSigns          map[int]rune
	BlameEnabled      bool
	BlameLines        map[int]BlameInfo
	BlameWidth        int
	VisualBlockInsert *VisualBlockInsertState

	OpenCount int
	rootDir   string
	Dirty     bool
}

func NewBuffer() *Buffer {
	lines := List[Line]{}
	lines.PushBack([]rune("\n"))
	b := &Buffer{
		Lines:       lines,
		IndentCh:    []rune{'\t'},
		Selection:   nil,
		Driver:      nil,
		Tx:          nil,
		Highlighter: nil,
		OpenCount:   1,
	}
	b.UndoRedo = NewUndoRedo(b)
	return b
}

func BufferReadFile(path string) (*Buffer, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := NewBuffer()
	buf.FilePath = path
	buf.Selection = nil
	buf.ResetLines()

	newLine := "\n"
	sc := bufio.NewScanner(file)
	i := 0
	for sc.Scan() {
		buf.Lines.PushBack([]rune(string(sc.Bytes()) + newLine))
		i++
	}
	if i == 0 {
		buf.Lines.PushBack([]rune(newLine))
	}

	return buf, nil
}

func BufferReloadFile(buf *Buffer) error {
	file, err := os.OpenFile(buf.FilePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	buf.Selection = nil
	buf.BlameEnabled = false
	buf.BlameLines = nil
	buf.BlameWidth = 0
	buf.Lines = List[Line]{}

	newLine := "\n"
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		buf.Lines.PushBack([]rune(string(sc.Bytes()) + newLine))
	}

	return nil
}

// ReloadBufferContent replaces the buffer's contents with the given text,
// recorded as a single undo/redo transaction. Use this instead of
// BufferReloadFile whenever the caller wants undo to work across the reload.
func ReloadBufferContent(ctx Context, content string) {
	buf := ctx.Buf

	if buf.TxStart() {
		defer buf.TxEnd()
	}

	// Delete everything: from (0,0) to the end of the last line (incl. its '\n').
	if buf.Lines.Len > 0 {
		lastLine := buf.Lines.Last()
		TextDelete(buf, &Selection{
			Start: Cursor{Line: 0, Char: 0},
			End:   Cursor{Line: buf.Lines.Len - 1, Char: len(lastLine.Value)},
		})
	}

	// Insert new content. Newline handling: TextInsert splits on '\n'.
	firstLine := buf.Lines.First()
	if firstLine == nil {
		// Buffer is empty — seed with an empty line first.
		buf.Lines.PushBack([]rune("\n"))
		firstLine = buf.Lines.First()
	}
	if content != "" {
		TextInsert(buf, firstLine, 0, content)
	}
}

func (buf *Buffer) SetMode(m Mode) {
	buf.mode = m
}

func (buf *Buffer) GetRootRir() string {
	if buf.rootDir == "" {
		buf.rootDir, _ = EditorInst.Projects.FindRoot(buf)
	}
	return buf.rootDir
}

func (b *Buffer) TxStart() (started bool) {
	if b.Tx != nil {
		return
	}

	b.Tx = NewTx(b)
	b.Tx.Start()
	return true
}

func (b *Buffer) TxEnd() {
	if b.Tx == nil {
		return
	}

	b.Tx.End()
	b.Tx = nil
}

func (b *Buffer) GetName() string {
	if len(b.FilePath) > 0 {
		return strings.Replace(b.FilePath, b.GetRootRir()+"/", "", 1)
	}
	return "[No Name]"
}

func (b *Buffer) Mode() Mode {
	return b.mode
}

func (b *Buffer) Save() error {
	f, err := os.Create(b.FilePath)
	if err != nil {
		return err
	}
	line := b.Lines.First()
	for line != nil {
		// temp check
		{
			count := 0
			for _, c := range line.Value {
				if c == '\n' {
					count++
				}
			}
			if count != 1 {
				EditorInst.LogMessage("wrong number of new lines")
				buf := EditorInst.BufferFindByFilePath("[Messages]", true)
				EditorInst.EnsureBufferIsVisible(buf)
			}
		}

		_, err := f.WriteString(string(line.Value))
		if err != nil {
			return err
		}
		line = line.Next()
	}
	return nil
}

func (b *Buffer) Append(s string) {
	// TODO: rewrite. use TextInsert as this messes up lsp and treesitter
	lines := strings.Split(s, "\n")
	// If the string ends with a newline, strings.Split produces a trailing empty string.
	// This would add a bogus empty line at the end of the buffer, which breaks git diff
	// and causes undo/redo to accumulate blank lines at the end of the file.
	if len(s) > 0 && s[len(s)-1] == '\n' {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		b.Lines.PushBack([]rune(line + "\n"))
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
	b.Lines = List[Line]{}
}

func (b *Buffer) CountLines() int {
	i := 0
	l := b.Lines.First()
	for l != nil {
		i++
		l = l.Next()
	}
	return i
}

func (b *Buffer) String() string {
	buf := bytes.NewBuffer(nil)
	line := b.Lines.First()
	for line != nil {
		buf.WriteString(string(line.Value))
		line = line.Next()
	}
	return buf.String()
}
