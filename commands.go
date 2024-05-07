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

func preserveCharPosition(buf *Buffer) {
	line := cursorToLine(buf)
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
	if e.activeBuffer.Cursor.Char > 0 {
		e.activeBuffer.Cursor.Char--
		e.activeBuffer.Cursor.PreserveCharPosition = e.activeBuffer.Cursor.Char
	}
}

func CmdCursorRight(e *Editor) {
	line := cursorToLine(e.activeBuffer)
	if e.activeBuffer.Cursor.Char < len(line.Data)-1 {
		e.activeBuffer.Cursor.Char++
		e.activeBuffer.Cursor.PreserveCharPosition = e.activeBuffer.Cursor.Char
	}
}

func CmdCursorLineUp(e *Editor) {
	if e.activeBuffer.Cursor.Line > 0 {
		e.activeBuffer.Cursor.Line--
		preserveCharPosition(e.activeBuffer)

		if e.activeBuffer.Cursor.Line < e.activeBuffer.ScrollOffset+3 {
			CmdScrollUp(e)
		}
	}
}

func CmdInsertMode(e *Editor) {
	e.activeBuffer.Mode = MODE_INSERT
}

func CmdNormalMode(e *Editor) {
	e.activeBuffer.Mode = MODE_NORMAL
}

func CmdCursorLineDown(e *Editor) {
	if e.activeBuffer.Cursor.Line < e.activeBuffer.Lines.Size-1 {
		e.activeBuffer.Cursor.Line++
		preserveCharPosition(e.activeBuffer)

		_, h := e.screen.Size()
		if e.activeBuffer.Cursor.Line-e.activeBuffer.ScrollOffset > h-3 {
			CmdScrollDown(e)
		}
	}
}

func CmdGotoLine0(e *Editor) {
	e.activeBuffer.Cursor.Line = 0
	e.activeBuffer.ScrollOffset = 0
	preserveCharPosition(e.activeBuffer)
}

func CmdGotoLineBegin(e *Editor) {
	e.activeBuffer.Cursor.Char = 0
}

func CmdGotoLineEnd(e *Editor) {
	line := cursorToLine(e.activeBuffer)
	e.activeBuffer.Cursor.Char = len(line.Data) - 1
}

func CmdForwardWord(e *Editor) {
	line := cursorToLine(e.activeBuffer)
	if e.activeBuffer.Cursor.Char < len(line.Data)-1 {
		e.activeBuffer.Cursor.Char++
		for e.activeBuffer.Cursor.Char < len(line.Data)-1 && !isSpecialChar(line.Data[e.activeBuffer.Cursor.Char]) {
			e.activeBuffer.Cursor.Char++
			e.activeBuffer.Cursor.PreserveCharPosition = e.activeBuffer.Cursor.Char
		}
	} else {
		CmdCursorLineDown(e)
		CmdGotoLineBegin(e)
	}
}

func CmdBackwardWord(e *Editor) {
	line := cursorToLine(e.activeBuffer)
	if e.activeBuffer.Cursor.Char > 0 {
		e.activeBuffer.Cursor.Char--
		for e.activeBuffer.Cursor.Char < len(line.Data)-1 && !isSpecialChar(line.Data[e.activeBuffer.Cursor.Char]) {
			e.activeBuffer.Cursor.Char--
			e.activeBuffer.Cursor.PreserveCharPosition = e.activeBuffer.Cursor.Char
		}
	} else {
		CmdCursorLineUp(e)
		CmdGotoLineEnd(e)
	}
}
