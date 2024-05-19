package mcwig

import (
	"unicode"
)

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

func lineByNum(buf *Buffer, num int) *Line {
	i := 0
	currentLine := buf.Lines.Head
	for currentLine != nil {
		if i == num {
			return currentLine
		}
		currentLine = currentLine.Next
		i++
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

func cursorGotoChar(buf *Buffer, ch int) {
	buf.Cursor.Char = ch
	buf.Cursor.PreserveCharPosition = buf.Cursor.Char
}

func lineJoinNext(buf *Buffer, line *Line) {
	next := line.Next
	if next == nil {
		return
	}

	line.Data = append(line.Data, next.Data...)
	buf.Lines.Delete(next)
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

func CmdInsertMode(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if len(line.Data) == 0 {
		CmdInsertModeAfter(e)
		return
	}

	e.ActiveBuffer.Mode = MODE_INSERT
}

func CmdVisualMode(e *Editor) {
	buf := e.ActiveBuffer
	buf.Selection = &Selection{
		Start: buf.Cursor,
		End:   buf.Cursor,
	}
	buf.Mode = MODE_VISUAL
}

func CmdInsertModeAfter(e *Editor) {
	e.ActiveBuffer.Cursor.Char++
	e.ActiveBuffer.Mode = MODE_INSERT
}

func CmdNormalMode(e *Editor) {
	buf := e.ActiveBuffer
	if buf.Mode == MODE_INSERT {
		CmdCursorLeft(e)
	}

	if buf.Mode == MODE_VISUAL {
		buf.Selection = nil
	}

	buf.Mode = MODE_NORMAL
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
	if e.ActiveBuffer.Cursor.Char >= len(line.Data) {
		CmdCursorLineDown(e)
		CmdCursorBeginningOfTheLine(e)
	} else {
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
	}
}

// TODO: fix select first word
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
	buf := e.ActiveBuffer
	line := cursorToLine(buf)

	if len(line.Data) == 0 {
		buf.Lines.DeleteByIndex(buf.Cursor.Line)
		CmdCursorLineUp(e)
		CmdAppendLine(e)
		return
	}

	if buf.Cursor.Char == 0 {
		CmdCursorLineUp(e)
		line = cursorToLine(buf)
		pos := len(line.Data)
		lineJoinNext(buf, line)
		cursorGotoChar(buf, pos)
		return
	}

	if buf.Cursor.Char >= len(line.Data) && len(line.Data) > 0 {
		line.Data = line.Data[:len(line.Data)-1]
		CmdCursorLeft(e)
		return
	}

	CmdCursorLeft(e)
	CmdDeleteCharForward(e)
}

func CmdAppendLine(e *Editor) {
	CmdGotoLineEnd(e)
	CmdInsertModeAfter(e)
}

func CmdNewLine(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	if len(line.Data) == 0 {
		buf.Lines.Insert(nil, buf.Cursor.Line)
	} else {
		tmpData := line.Data[buf.Cursor.Char:]
		line.Data = line.Data[:buf.Cursor.Char]
		buf.Lines.Insert(tmpData, buf.Cursor.Line+1)
	}
	CmdCursorLineDown(e)
	CmdCursorBeginningOfTheLine(e)
}

func CmdLineOpenBelow(e *Editor) {
	CmdAppendLine(e)
	CmdNewLine(e)
	CmdInsertModeAfter(e)
}

func CmdLineOpenAbove(e *Editor) {
	CmdCursorLineUp(e)
	CmdLineOpenBelow(e)
}

func CmdDeleteLine(e *Editor) {
	buf := e.ActiveBuffer
	buf.Lines.DeleteByIndex(buf.Cursor.Line)

	if buf.Cursor.Line >= buf.Lines.Size {
		CmdCursorLineUp(e)
		CmdCursorBeginningOfTheLine(e)
	}
}

func CmdChangeLine(e *Editor) {
	buf := e.ActiveBuffer
	line := cursorToLine(buf)
	line.Data = nil
	CmdInsertModeAfter(e)
}

// FIXME: delete last char on last line
func CmdSelectinDelete(e *Editor) {
	buf := e.ActiveBuffer
	if buf.Selection == nil {
		return
	}

	curStart := buf.Selection.Start
	curEnd := buf.Selection.End

	if curStart.Line > curEnd.Line {
		curStart, curEnd = curEnd, curStart
	}

	lineNum := curStart.Line
	lineStart := lineByNum(buf, curStart.Line)
	lineEnd := lineByNum(buf, curEnd.Line)

	if lineNum == curStart.Line && lineNum == curEnd.Line {
		if curStart.Char > curEnd.Char {
			curStart, curEnd = curEnd, curStart
		}
		lineStart.Data = append(lineStart.Data[:curStart.Char], lineStart.Data[curEnd.Char+1:]...)
	} else {
		lineStart.Next = lineEnd
		lineEnd.Prev = lineStart
		lineStart.Data = lineStart.Data[:curStart.Char]
		lineEnd.Data = lineEnd.Data[curEnd.Char+1:]
		if len(lineEnd.Data) == 0 {
			buf.Lines.Delete(lineEnd)
		}
		lineJoinNext(buf, lineStart)
	}

	// TODO: fix panic on last line delete
	buf.Cursor.Line = curStart.Line
	if lineStart != nil && curStart.Char < len(lineStart.Data) {
		cursorGotoChar(buf, curStart.Char)
	} else {
		CmdGotoLineEnd(e)
	}
	CmdNormalMode(e)
}

func WithSelection(e *Editor, fn func(*Editor)) func(*Editor) {
	return func(e *Editor) {
		fn(e)
		buf := e.ActiveBuffer
		buf.Selection.End = buf.Cursor
	}
}

func WithSelectionToChar(e *Editor, fn func(*Editor, string)) func(*Editor, string) {
	return func(e *Editor, ch string) {
		fn(e, ch)
		buf := e.ActiveBuffer
		buf.Selection.End = buf.Cursor
	}
}
