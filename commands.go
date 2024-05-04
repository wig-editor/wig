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
	if buf.Cursor.PreserveCharPosition > len(line.Data) {
		buf.Cursor.Char = len(line.Data) - 1
	} else {
		buf.Cursor.Char = buf.Cursor.PreserveCharPosition
	}
}

func CmdScrollUp(e *Editor) {
	if e.activeBuffer.ScrollOffset > 0 {
		e.activeBuffer.ScrollOffset--
	}
}

func CmdScrollDown(e *Editor) {
	if e.activeBuffer.ScrollOffset < e.activeBuffer.Lines.Size-3 {
		e.activeBuffer.ScrollOffset++
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
	}
}

func CmdCursorLineDown(e *Editor) {
	if e.activeBuffer.Cursor.Line < e.activeBuffer.Lines.Size {
		e.activeBuffer.Cursor.Line++
		preserveCharPosition(e.activeBuffer)
	}
}
