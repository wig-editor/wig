package mcwig

import "unicode"

// TODO: check performance on big files. cache pointer?
func cursorToLine(buf *Buffer) *Line {
	num := 0
	currentLine := buf.Lines.Head
	for currentLine != nil {
		if buf.Cursor.Line == num {
			return currentLine
		}
		currentLine = currentLine.Next
		num++
	}
	return currentLine
}

func restoreCharPosition(buf *Buffer) {
	line := cursorToLine(buf)

	if len(line.Data) == 0 {
		buf.Cursor.Char = 0
		return
	}

	if buf.Cursor.PreserveCharPosition >= len(line.Data) {
		buf.Cursor.Char = len(line.Data) - 1
	} else {
		buf.Cursor.Char = buf.Cursor.PreserveCharPosition
	}
}

func isSpecialChar(c rune) bool {
	specialChars := []rune(" ,.()[]{}<>:;+*/-=~!@#$%^&|?`\"")
	for _, char := range specialChars {
		if c == char {
			return true
		}
	}
	return false
}

func CmdScrollUp(e *Editor) {
	if e.ActiveBuffer.ScrollOffset > 0 {
		e.ActiveBuffer.ScrollOffset--

		_, h := e.Screen.Size()
		if e.ActiveBuffer.Cursor.Line > e.ActiveBuffer.ScrollOffset+h-3 {
			CmdCursorLineUp(e)
		}
	}
}

func CmdScrollDown(e *Editor) {
	if e.ActiveBuffer.ScrollOffset < e.ActiveBuffer.Lines.Size-3 {
		e.ActiveBuffer.ScrollOffset++

		if e.ActiveBuffer.Cursor.Line <= e.ActiveBuffer.ScrollOffset+3 {
			CmdCursorLineDown(e)
		}
	}
}

func CmdCursorLeft(e *Editor) {
	buf := e.ActiveBuffer
	if buf.Cursor.Char > 0 {
		buf.Cursor.Char--
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
	}
}

func CmdCursorRight(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if buf.Cursor.Char < len(line.Data)-1 {
		buf.Cursor.Char++
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
	}
}

func CmdCursorLineUp(e *Editor) {
	if e.ActiveBuffer.Cursor.Line > 0 {
		e.ActiveBuffer.Cursor.Line--
		restoreCharPosition(e.ActiveBuffer)

		if e.ActiveBuffer.Cursor.Line < e.ActiveBuffer.ScrollOffset+3 {
			CmdScrollUp(e)
		}
	}
}

func CmdCursorBeginningOfTheLine(e *Editor) {
	e.ActiveBuffer.Cursor.Char = 0
	e.ActiveBuffer.Cursor.PreserveCharPosition = 0
}

func CmdCursorFirstNonBlank(e *Editor) {
	CmdCursorBeginningOfTheLine(e)
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if len(line.Data) == 0 {
		return
	}
	for _, c := range line.Data {
		if unicode.IsSpace(c) {
			CmdCursorRight(e)
		} else {
			break
		}
	}
}

func CmdCursorLineDown(e *Editor) {
	if e.ActiveBuffer.Cursor.Line < e.ActiveBuffer.Lines.Size-1 {
		e.ActiveBuffer.Cursor.Line++
		restoreCharPosition(e.ActiveBuffer)

		_, h := e.Screen.Size()
		if e.ActiveBuffer.Cursor.Line-e.ActiveBuffer.ScrollOffset > h-3 {
			CmdScrollDown(e)
		}
	}
}

func CmdInsertMode(e *Editor) {
	e.ActiveBuffer.Mode = MODE_INSERT
}

func CmdNormalMode(e *Editor) {
	e.ActiveBuffer.Mode = MODE_NORMAL
}

func CmdGotoLine0(e *Editor) {
	e.ActiveBuffer.Cursor.Line = 0
	e.ActiveBuffer.ScrollOffset = 0
	restoreCharPosition(e.ActiveBuffer)
}

func CmdGotoLineEnd(e *Editor) {
	line := cursorToLine(e.ActiveBuffer)
	e.ActiveBuffer.Cursor.Char = len(line.Data) - 1
}

func CmdForwardWord(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if e.ActiveBuffer.Cursor.Char < len(line.Data) {
		for {
			if e.ActiveBuffer.Cursor.Char >= len(line.Data)-1 {
				CmdCursorLineDown(e)
				CmdCursorBeginningOfTheLine(e)
				CmdCursorFirstNonBlank(e)
				break
			}

			CmdCursorRight(e)

			if isSpecialChar(line.Data[e.ActiveBuffer.Cursor.Char]) {
				break
			}
		}
	} else {
		CmdCursorLineDown(e)
		CmdCursorBeginningOfTheLine(e)
	}
}

func CmdForwardChar(e *Editor, ch string) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if len(line.Data) == 0 {
		return
	}

	for i := buf.Cursor.Char + 1; i < len(line.Data); i++ {
		if string(line.Data[i]) == ch {
			buf.Cursor.Char = i
			buf.Cursor.PreserveCharPosition = i
			break
		}
	}
}

func CmdBackwardChar(e *Editor, ch string) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if len(line.Data) == 0 {
		return
	}

	for i := buf.Cursor.Char - 1; i >= 0; i-- {
		if string(line.Data[i]) == ch {
			buf.Cursor.Char = i
			buf.Cursor.PreserveCharPosition = i
			break
		}
	}
}

func CmdBackwardWord(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if e.ActiveBuffer.Cursor.Char > 0 {
		for {
			if e.ActiveBuffer.Cursor.Char == 0 {
				CmdCursorLineUp(e)
				CmdGotoLineEnd(e)
				break
			}

			CmdCursorLeft(e)

			if isSpecialChar(line.Data[e.ActiveBuffer.Cursor.Char]) {
				break
			}
		}
	} else {
		CmdCursorLineUp(e)
		CmdGotoLineEnd(e)
	}
}

func CmdDeleteCharForward(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if len(line.Data) == 0 {
		return
	}
	line.Data = append(line.Data[:buf.Cursor.Char], line.Data[buf.Cursor.Char+1:]...)
	if buf.Cursor.Char >= len(line.Data) {
		CmdCursorLeft(e)
	}
}

func CmdDeleteCharBackward(e *Editor) {
	CmdCursorLeft(e)
	CmdDeleteCharForward(e)
}
