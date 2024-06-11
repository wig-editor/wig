package mcwig

import "unicode"

func restoreCharPosition(buf *Buffer) {
	line := CursorLine(buf)
	if line == nil {
		buf.Cursor.Char = 0
		return
	}

	if len(line.Value) == 0 {
		buf.Cursor.Char = 0
		return
	}

	if buf.Cursor.PreserveCharPosition >= len(line.Value) {
		buf.Cursor.Char = len(line.Value) - 1
	} else {
		buf.Cursor.Char = buf.Cursor.PreserveCharPosition
	}
}

func cursorGotoChar(buf *Buffer, ch int) {
	buf.Cursor.Char = ch
	buf.Cursor.PreserveCharPosition = buf.Cursor.Char
}

func lineJoinNext(buf *Buffer, line *Element[Line]) {
	next := line.Next()
	if next == nil {
		return
	}

	line.Value = append(line.Value, next.Value...)
	buf.Lines.Remove(next)
}

func CursorInc(buf *Buffer) (moved bool) {
	line := CursorLine(buf)
	if buf.Cursor.Char < len(line.Value)-1 {
		buf.Cursor.Char++
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		return true
	}

	if line.Next() != nil {
		buf.Cursor.Char = 0
		buf.Cursor.Line++
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		return true
	}

	return false
}

func CursorDec(buf *Buffer) (moved bool) {
	if buf.Cursor.Char > 0 {
		buf.Cursor.Char--
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		return true
	}

	line := CursorLine(buf)
	if line.Prev() != nil {
		chLen := len(line.Prev().Value) - 1
		if chLen < 0 {
			chLen = 0
		}

		buf.Cursor.Char = chLen
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		buf.Cursor.Line--
		return true
	}

	return false
}

func CursorLine(buf *Buffer) *Element[Line] {
	num := 0
	currentLine := buf.Lines.First()
	for currentLine != nil {
		if buf.Cursor.Line == num {
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

// class of char under cursor
type chClass int

const (
	chWhitespace chClass = iota
	chPunct
	chWord
)

func CursorChClass(buf *Buffer) chClass {
	line := CursorLine(buf)

	if len(line.Value) == 0 {
		return chWhitespace
	}

	chLen := buf.Cursor.Char
	if chLen > len(line.Value)-1 {
		chLen = len(line.Value) - 1
	}

	r := line.Value[chLen]

	if unicode.IsSpace(r) {
		return chWhitespace
	}

	if unicode.IsPunct(r) || unicode.IsSymbol(r) {
		return chPunct
	}

	return chWord
}
