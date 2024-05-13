package mcwig

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
	if e.activeBuffer.ScrollOffset > 0 {
		e.activeBuffer.ScrollOffset--

		_, h := e.screen.Size()
		if e.activeBuffer.Cursor.Line > e.activeBuffer.ScrollOffset+h-3 {
			CmdCursorLineUp(e)
		}
	}
}

func CmdScrollDown(e *Editor) {
	if e.activeBuffer.ScrollOffset < e.activeBuffer.Lines.Size-3 {
		e.activeBuffer.ScrollOffset++

		if e.activeBuffer.Cursor.Line <= e.activeBuffer.ScrollOffset+3 {
			CmdCursorLineDown(e)
		}
	}
}

func CmdCursorLeft(e *Editor) {
	buf := e.activeBuffer
	if buf.Cursor.Char > 0 {
		buf.Cursor.Char--
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
	}
}

func CmdCursorRight(e *Editor) {
	buf := e.activeBuffer
	line := cursorToLine(buf)
	if buf.Cursor.Char < len(line.Data)-1 {
		buf.Cursor.Char++
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
	}
}

func CmdCursorLineUp(e *Editor) {
	if e.activeBuffer.Cursor.Line > 0 {
		e.activeBuffer.Cursor.Line--
		restoreCharPosition(e.activeBuffer)

		if e.activeBuffer.Cursor.Line < e.activeBuffer.ScrollOffset+3 {
			CmdScrollUp(e)
		}
	}
}

func CmdCursorBeginningOfTheLine(e *Editor) {
	e.activeBuffer.Cursor.Char = 0
	e.activeBuffer.Cursor.PreserveCharPosition = 0
}

func CmdCursorLineDown(e *Editor) {
	if e.activeBuffer.Cursor.Line < e.activeBuffer.Lines.Size-1 {
		e.activeBuffer.Cursor.Line++
		restoreCharPosition(e.activeBuffer)

		_, h := e.screen.Size()
		if e.activeBuffer.Cursor.Line-e.activeBuffer.ScrollOffset > h-3 {
			CmdScrollDown(e)
		}
	}
}

func CmdInsertMode(e *Editor) {
	e.activeBuffer.Mode = MODE_INSERT
}

func CmdNormalMode(e *Editor) {
	e.activeBuffer.Mode = MODE_NORMAL
}

func CmdGotoLine0(e *Editor) {
	e.activeBuffer.Cursor.Line = 0
	e.activeBuffer.ScrollOffset = 0
	restoreCharPosition(e.activeBuffer)
}

func CmdGotoLineEnd(e *Editor) {
	line := cursorToLine(e.activeBuffer)
	e.activeBuffer.Cursor.Char = len(line.Data) - 1
}

func CmdForwardWord(e *Editor) {
	buf := e.activeBuffer
	line := cursorToLine(buf)
	if e.activeBuffer.Cursor.Char < len(line.Data) {
		for {
			if e.activeBuffer.Cursor.Char >= len(line.Data)-1 {
				CmdCursorLineDown(e)
				CmdCursorBeginningOfTheLine(e)
				break
			}

			CmdCursorRight(e)

			if isSpecialChar(line.Data[e.activeBuffer.Cursor.Char]) {
				break
			}
		}
	} else {
		CmdCursorLineDown(e)
		CmdCursorBeginningOfTheLine(e)
	}
}

func CmdBackwardWord(e *Editor) {
	buf := e.activeBuffer
	line := cursorToLine(buf)
	if e.activeBuffer.Cursor.Char > 0 {
		for {
			if e.activeBuffer.Cursor.Char == 0 {
				CmdCursorLineUp(e)
				CmdGotoLineEnd(e)
				break
			}

			CmdCursorLeft(e)

			if isSpecialChar(line.Data[e.activeBuffer.Cursor.Char]) {
				break
			}
		}
	} else {
		CmdCursorLineUp(e)
		CmdGotoLineEnd(e)
	}
}

func CmdDeleteCharForward(e *Editor) {
	buf := e.activeBuffer
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
