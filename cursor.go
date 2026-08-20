package wig

import (
	"unicode"

	"github.com/mattn/go-runewidth"
)

type Cursor struct {
	Line                 int
	Char                 int
	PreserveCharPosition int
	ScrollOffset         int
}

type Location struct {
	Text     string
	FilePath string
	Line     int
	Char     int
}

// restoreCharPosition places the cursor on the current line at the rune
// index corresponding to the sticky visual column (cur.PreserveCharPosition),
// so vertical movement (j/k, scrolling) lines up on-screen even when lines
// have different tab/space leading whitespace instead of just reusing the
// same rune index across lines of differing tab counts.
func restoreCharPosition(buf *Buffer, cur *Cursor) {
	line := CursorLine(buf, cur)
	if line == nil {
		cur.Char = 0
		return
	}
	if len(line.Value) == 0 {
		cur.Char = 0
		return
	}
	maxChar := len(line.Value) - 1
	idx := RuneIndexFromVisualCol(line.Value, cur.PreserveCharPosition)
	if idx > maxChar {
		idx = maxChar
	}
	cur.Char = idx
}
func CursorInc(buf *Buffer, cur *Cursor) (moved bool) {
	line := CursorLine(buf, cur)
	if line == nil {
		return false
	}
	if cur.Char < len(line.Value)-1 {
		cur.Char++
		cur.PreserveCharPosition = VisualCol(line.Value, cur.Char)
		return true
	}
	if line.Next() != nil {
		cur.Char = 0
		cur.Line++
		cur.PreserveCharPosition = 0
		return true
	}
	return false
}
func CursorDec(buf *Buffer, cur *Cursor) (moved bool) {
	if cur.Char > 0 {
		cur.Char--
		if line := CursorLine(buf, cur); line != nil {
			cur.PreserveCharPosition = VisualCol(line.Value, cur.Char)
		}
		return true
	}
	line := CursorLine(buf, cur)
	if line == nil {
		return false
	}
	if line.Prev() != nil {
		prevLine := line.Prev()
		chLen := max(len(prevLine.Value)-1, 0)
		cur.Char = chLen
		cur.PreserveCharPosition = VisualCol(prevLine.Value, chLen)
		cur.Line--
		return true
	}
	return false
}
func CursorLine(buf *Buffer, cur *Cursor) *Element[Line] {
	num := 0
	currentLine := buf.Lines.First()
	for currentLine != nil {
		if cur.Line == num {
			return currentLine
		}
		currentLine = currentLine.Next()
		num++
	}
	return currentLine
}

func CursorLineByNum(buf *Buffer, num int) *Element[Line] {
	i := 0
	currentLine := buf.Lines.First()
	for currentLine != nil {
		if i == num {
			return currentLine
		}
		currentLine = currentLine.Next()
		i++
	}

	return currentLine
}

func CursorNumByLine(buf *Buffer, lookie *Element[Line]) int {
	i := 0
	currentLine := buf.Lines.First()
	for currentLine != nil {
		if currentLine == lookie {
			return i
		}
		currentLine = currentLine.Next()
		i++
	}

	return 0
}

func ContextCursorGet(ctx Context) *Cursor {
	win := ctx.Win
	if win == nil {
		win = ctx.Editor.ActiveWindow()
	}
	return WindowCursorGet(win, ctx.Buf)
}

func CursorGet(editor *Editor, buf *Buffer) *Cursor {
	win := editor.ActiveWindow()
	return WindowCursorGet(win, buf)
}

func WindowCursorGet(win *Window, buf *Buffer) *Cursor {
	cur, ok := win.cursors[buf]
	if ok {
		return cur
	}

	cur = &Cursor{}
	win.cursors[buf] = cur
	return cur
}

func WindowCursorSet(win *Window, buf *Buffer, cur *Cursor) {
	win.cursors[buf] = cur
}

// VisualCol calculates the visual screen column of a given rune index.
func VisualCol(line []rune, char int) int {
	col := 0
	for i := 0; i < char && i < len(line); i++ {
		if line[i] == '\t' {
			col += 4
		} else if line[i] == '\n' {
			// skip
		} else {
			col += runewidth.RuneWidth(line[i])
		}
	}
	return col
}

// RuneIndexFromVisualCol finds the rune index that corresponds to a given visual screen column.
func RuneIndexFromVisualCol(line []rune, visCol int) int {
	col := 0
	for i, r := range line {
		if col >= visCol {
			return i
		}
		if r == '\t' {
			col += 4
		} else if r == '\n' {
			// skip
		} else {
			col += runewidth.RuneWidth(r)
		}
	}
	return len(line)
}

// class of char under cursor
type chClass int

const (
	chWhitespace chClass = iota
	chPunct
	chWord
)

func CursorChClass(buf *Buffer, cur *Cursor) chClass {
	line := CursorLine(buf, cur)

	if line == nil || len(line.Value) == 0 {
		return chWhitespace
	}

	chLen := cur.Char
	if chLen > len(line.Value)-1 {
		chLen = len(line.Value) - 1
	}

	return getChClass(line.Value[chLen])
}

// Returns char under the cursor.
func CursorChar(buf *Buffer, cur *Cursor) rune {
	line := CursorLine(buf, cur)

	if line == nil || line.Value.IsEmpty() {
		return -1
	}

	return line.Value[cur.Char]
}

func getChClass(r rune) chClass {
	if unicode.IsSpace(r) {
		return chWhitespace
	}

	if r == '_' {
		return chWord
	}

	if unicode.IsPunct(r) || unicode.IsSymbol(r) {
		return chPunct
	}

	return chWord
}
