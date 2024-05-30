package mcwig

import (
	"unicode"
)

func cursorToLine(buf *Buffer) *Element[Line] {
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

func lineByNum(buf *Buffer, num int) *Element[Line] {
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

func restoreCharPosition(buf *Buffer) {
	line := cursorToLine(buf)
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

func isSpecialChar(c rune) bool {
	specialChars := []rune(",.()[]{}<>:;+*/-=~!@#$%^&|?`\"")
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

func lineJoinNext(buf *Buffer, line *Element[Line]) {
	next := line.Next()
	if next == nil {
		return
	}

	line.Value = append(line.Value, next.Value...)
	buf.Lines.Remove(next)
}

func do(e *Editor, fn func(buf *Buffer, line *Element[Line])) {
	buf := e.ActiveBuffer()
	if buf != nil {
		line := cursorToLine(buf)
		fn(buf, line)
	}
}

func CmdJoinNextLine(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		CmdGotoLineEnd(e)
		lineJoinNext(buf, line)
	})
}

func CmdScrollUp(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.ScrollOffset > 0 {
			buf.ScrollOffset--

			_, h := e.View.Size()
			if buf.Cursor.Line > buf.ScrollOffset+h-3 {
				CmdCursorLineUp(e)
			}
		}
	})
}

func CmdScrollDown(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.ScrollOffset < buf.Lines.Len-3 {
			buf.ScrollOffset++

			if buf.Cursor.Line <= buf.ScrollOffset+3 {
				CmdCursorLineDown(e)
			}
		}
	})
}

func CmdCursorLeft(e *Editor) {
	do(e, func(buf *Buffer, _ *Element[Line]) {
		if buf.Cursor.Char > 0 {
			buf.Cursor.Char--
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	})
}

func CmdCursorRight(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Char < len(line.Value)-1 {
			buf.Cursor.Char++
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	})
}

func CmdCursorLineUp(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Line > 0 {
			buf.Cursor.Line--
			restoreCharPosition(buf)

			if buf.Cursor.Line < buf.ScrollOffset+3 {
				CmdScrollUp(e)
			}
		}
	})
}

func CmdCursorLineDown(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Line < buf.Lines.Len-1 {
			buf.Cursor.Line++
			restoreCharPosition(buf)

			_, h := e.View.Size()
			if buf.Cursor.Line-buf.ScrollOffset > h-3 {
				CmdScrollDown(e)
			}
		}
	})
}

func CmdCursorBeginningOfTheLine(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Char = 0
		buf.Cursor.PreserveCharPosition = 0
	})
}

func CmdCursorFirstNonBlank(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		CmdCursorBeginningOfTheLine(e)
		if len(line.Value) == 0 {
			return
		}
		for _, c := range line.Value {
			if unicode.IsSpace(c) {
				CmdCursorRight(e)
			} else {
				break
			}
		}
	})
}

func CmdInsertMode(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			CmdInsertModeAfter(e)
			return
		}

		buf.Mode = MODE_INSERT
	})
}

func CmdVisualMode(e *Editor) {
	do(e, func(buf *Buffer, _ *Element[Line]) {
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   buf.Cursor,
		}
		buf.Mode = MODE_VISUAL
	})
}

func CmdInsertModeAfter(e *Editor) {
	do(e, func(buf *Buffer, _ *Element[Line]) {
		buf.Cursor.Char++
		buf.Mode = MODE_INSERT
	})
}

func CmdNormalMode(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Mode == MODE_INSERT {
			CmdCursorLeft(e)
			if buf.Cursor.Char >= len(line.Value) {
				CmdGotoLineEnd(e)
			}
		}

		buf.Selection = nil
		buf.Mode = MODE_NORMAL
	})
}

func CmdGotoLine0(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Line = 0
		buf.ScrollOffset = 0
		restoreCharPosition(buf)
	})
}

func CmdGotoLineEnd(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Char = len(line.Value) - 1
	})
}

func CmdForwardWord(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		checkEOF := func(line *Element[Line]) bool {
			if buf.Cursor.Char >= len(line.Value) || line.Value.IsEmpty() {
				CmdCursorLineDown(e)
				CmdCursorFirstNonBlank(e)
				return true
			}
			return false
		}
		if checkEOF(line) {
			return
		}

		exitOn := isSpecialChar

		ch := line.Value[buf.Cursor.Char]

		for {
			if checkEOF(line) {
				return
			}

			if buf.Cursor.Char >= len(line.Value)-1 {
				CmdCursorLineDown(e)
				CmdCursorFirstNonBlank(e)
				line = cursorToLine(buf)
				if line.Value.IsEmpty() {
					continue
				}
				break
			}

			if unicode.IsSpace(ch) || isSpecialChar(ch) {
				exitOn = func(ch rune) bool {
					return !unicode.IsSpace(ch)
				}
			}

			CmdCursorRight(e)

			ch = line.Value[buf.Cursor.Char]

			if exitOn(ch) {
				break
			}
		}
	})
}

func CmdBackwardWord(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Char == 0 && buf.Cursor.Line == 0 {
			return
		}

		if buf.Cursor.Char == 0 || line.Value.IsEmpty() {
			CmdCursorLineUp(e)
			CmdGotoLineEnd(e)
			return
		}

		CmdCursorLeft(e)

		for {
			line = cursorToLine(buf)
			ch := line.Value[buf.Cursor.Char]

			if buf.Cursor.Char == 0 {
				if unicode.IsSpace(ch) {
					CmdCursorLineUp(e)
					CmdGotoLineEnd(e)
				}
				break
			}

			if unicode.IsSpace(ch) {
				CmdCursorLeft(e)
				continue
			}

			if isSpecialChar(ch) {
				break
			}

			prevCh := line.Value[buf.Cursor.Char-1]
			if unicode.IsSpace(prevCh) || isSpecialChar(prevCh) {
				break
			}

			CmdCursorLeft(e)
		}
	})
}

func CmdForwardChar(e *Editor, ch string) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			return
		}

		for i := buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if string(line.Value[i]) == ch {
				buf.Cursor.Char = i
				buf.Cursor.PreserveCharPosition = i
				break
			}
		}
	})
}

func CmdBackwardChar(e *Editor, ch string) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			return
		}

		for i := buf.Cursor.Char - 1; i >= 0; i-- {
			if string(line.Value[i]) == ch {
				buf.Cursor.Char = i
				buf.Cursor.PreserveCharPosition = i
				break
			}
		}
	})
}

func CmdDeleteCharForward(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			CmdJoinNextLine(e)
			CmdCursorBeginningOfTheLine(e)
			return
		}
		line.Value = append(line.Value[:buf.Cursor.Char], line.Value[buf.Cursor.Char+1:]...)
		if buf.Cursor.Char >= len(line.Value) {
			CmdCursorLeft(e)
		}
	})
}

func CmdDeleteCharBackward(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			buf.Lines.Remove(line)
			CmdCursorLineUp(e)
			CmdAppendLine(e)
			return
		}

		if buf.Cursor.Char == 0 {
			CmdCursorLineUp(e)
			line = cursorToLine(buf)
			pos := len(line.Value)
			lineJoinNext(buf, line)
			cursorGotoChar(buf, pos)
			return
		}

		if buf.Cursor.Char >= len(line.Value) && len(line.Value) > 0 {
			line.Value = line.Value[:len(line.Value)-1]
			CmdCursorLeft(e)
			return
		}

		CmdCursorLeft(e)
		CmdDeleteCharForward(e)
	})
}

func CmdAppendLine(e *Editor) {
	CmdGotoLineEnd(e)
	CmdInsertModeAfter(e)
}

func CmdNewLine(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		// EOL
		if (buf.Cursor.Char) >= len(line.Value) {
			buf.Lines.insertValueAfter(Line{}, line)
			buf.Cursor.Line++
			buf.Cursor.Char = 1
			buf.Cursor.PreserveCharPosition = 0
			return
		}

		// split line
		tmpData := make([]rune, len(line.Value[buf.Cursor.Char:]))
		copy(tmpData, line.Value[buf.Cursor.Char:])
		line.Value = line.Value[:buf.Cursor.Char]
		buf.Lines.insertValueAfter(tmpData, line)
		CmdCursorLineDown(e)
		CmdCursorBeginningOfTheLine(e)
	})
}

func CmdLineOpenBelow(e *Editor) {
	CmdAppendLine(e)
	CmdNewLine(e)
	CmdInsertModeAfter(e)
}

func CmdLineOpenAbove(e *Editor) {
	do(e, func(buf *Buffer, _ *Element[Line]) {
		if buf.Cursor.Line == 0 {
			buf.Lines.PushFront(Line{})
			CmdCursorBeginningOfTheLine(e)
			CmdInsertModeAfter(e)
			return
		}
		CmdCursorLineUp(e)
		CmdLineOpenBelow(e)
	})
}

func CmdDeleteLine(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if line == nil {
			return
		}

		buf.Lines.Remove(line)
		if buf.Cursor.Line >= buf.Lines.Len {
			CmdCursorLineUp(e)
			CmdCursorBeginningOfTheLine(e)
		}

		if buf.Lines.Len == 0 {
			buf.Lines.PushFront(Line{})
		}

		restoreCharPosition(buf)
	})
}

func CmdChangeLine(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		line.Value = nil
		CmdInsertModeAfter(e)
	})
}

func CmdSelectinDelete(e *Editor) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			CmdNormalMode(e)
		}()
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
			lineStart.Value = append(lineStart.Value[:curStart.Char], lineStart.Value[curEnd.Char+1:]...)

			if len(lineStart.Value) == 0 {
				buf.Lines.Remove(lineStart)
				CmdCursorBeginningOfTheLine(e)
				return
			}

			cursorGotoChar(buf, curStart.Char)
		} else {
			// delete all lines between start and end line
			for lineStart.Next() != lineEnd {
				buf.Lines.Remove(lineStart.Next())
			}

			lineStart.Value = lineStart.Value[:curStart.Char]

			if curEnd.Char+1 <= len(lineEnd.Value) {
				lineEnd.Value = lineEnd.Value[curEnd.Char+1:]
			}

			if len(lineEnd.Value) == 0 {
				buf.Lines.Remove(lineEnd)
			}

			lineJoinNext(buf, lineStart)

			buf.Cursor.Line = curStart.Line
			if lineStart != nil && curStart.Char < len(lineStart.Value) {
				cursorGotoChar(buf, curStart.Char)
			} else {
				CmdGotoLineEnd(e)
			}
		}
	})
}

func CmdSaveFile(e *Editor) {
}

func CmdWindowVSplit(e *Editor) {
	do(e, func(buf *Buffer, _ *Element[Line]) {
		e.Windows = append(e.Windows, &Window{Buffer: buf})
	})
}

func CmdWindowNext(e *Editor) {
	curWin := e.activeWindow
	idx := 0
	for i, w := range e.Windows {
		if w == curWin {
			idx = i + 1
			break
		}
	}

	if idx >= len(e.Windows) {
		idx = 0
	}

	e.activeWindow = e.Windows[idx]
}

func CmdExit(e *Editor) {
	e.ExitCh <- 1
}

func WithSelection(fn func(*Editor)) func(*Editor) {
	return func(e *Editor) {
		fn(e)
		buf := e.ActiveBuffer()
		buf.Selection.End = buf.Cursor
	}
}

func WithSelectionToChar(fn func(*Editor, string)) func(*Editor, string) {
	return func(e *Editor, ch string) {
		fn(e, ch)
		buf := e.ActiveBuffer()
		buf.Selection.End = buf.Cursor
	}
}
