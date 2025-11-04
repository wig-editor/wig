package wig

import "unicode"

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

	if cur.PreserveCharPosition >= len(line.Value) {
		cur.Char = len(line.Value) - 1
	} else {
		cur.Char = cur.PreserveCharPosition
	}
}

func CursorInc(buf *Buffer, cur *Cursor) (moved bool) {
	line := CursorLine(buf, cur)
	if cur.Char < len(line.Value)-1 {
		cur.Char++
		cur.PreserveCharPosition = cur.Char
		return true
	}

	if line.Next() != nil {
		cur.Char = 0
		cur.Line++
		cur.PreserveCharPosition = cur.Char
		return true
	}

	return false
}

func CursorDec(buf *Buffer, cur *Cursor) (moved bool) {
	if cur.Char > 0 {
		cur.Char--
		cur.PreserveCharPosition = cur.Char
		return true
	}

	line := CursorLine(buf, cur)
	if line.Prev() != nil {
		chLen := max(len(line.Prev().Value)-1, 0)
		cur.Char = chLen
		cur.PreserveCharPosition = cur.Char
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
	win := ctx.Editor.ActiveWindow()
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

// class of char under cursor
type chClass int

const (
	chWhitespace chClass = iota
	chPunct
	chWord
)

func CursorChClass(buf *Buffer, cur *Cursor) chClass {
	line := CursorLine(buf, cur)

	if len(line.Value) == 0 {
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

	if line.Value.IsEmpty() {
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

