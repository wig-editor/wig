package mcwig

import (
	"fmt"
	"strings"
	"unicode"
)

func lineJoinNext(buf *Buffer, line *Element[Line]) {
	next := line.Next()
	if next == nil {
		return
	}

	line.Value = append(line.Value, next.Value...)
	buf.Lines.Remove(next)
}

func Do(e *Editor, fn func(buf *Buffer, line *Element[Line])) {
	buf := e.ActiveBuffer()
	if buf == nil {
		return
	}

	fn(buf, CursorLine(buf))
}

func CmdJoinNextLine(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		CmdGotoLineEnd(e)
		lineJoinNext(buf, line)
	})
}

func CmdScrollUp(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.ScrollOffset < buf.Lines.Len-3 {
			buf.ScrollOffset++

			if buf.Cursor.Line <= buf.ScrollOffset+3 {
				CmdCursorLineDown(e)
			}
		}
	})
}

func CmdCursorLeft(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		if buf.Cursor.Char > 0 {
			buf.Cursor.Char--
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	})
}

func CmdCursorRight(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Char < len(line.Value) {
			buf.Cursor.Char++
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	})
}

func CmdCursorLineUp(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
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
	Do(e, func(buf *Buffer, _ *Element[Line]) {
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Char = 0
		buf.Cursor.PreserveCharPosition = 0
	})
}

func CmdCursorFirstNonBlank(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			CmdInsertModeAfter(e)
			return
		}

		buf.SetMode(MODE_INSERT)
	})
}

func CmdVisualMode(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		SelectionStart(buf)
		buf.SetMode(MODE_VISUAL)
	})
}

func CmdVisualLineMode(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		buf.Selection.Start.Char = 0
		buf.Selection.End.Char = len(line.Value) - 1
		buf.SetMode(MODE_VISUAL_LINE)
	})
}

func CmdInsertModeAfter(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		buf.Cursor.Char++
		buf.SetMode(MODE_INSERT)
	})
}

func CmdNormalMode(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Mode() == MODE_INSERT {
			CmdCursorLeft(e)
			if buf.Cursor.Char >= len(line.Value) {
				CmdGotoLineEnd(e)
			}
		}
		buf.SetMode(MODE_NORMAL)
		buf.Selection = nil
	})
}

func CmdGotoLine0(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Line = 0
		buf.ScrollOffset = 0
		restoreCharPosition(buf)

		if e.Keys.GetTimes() > 1 {
			ln := e.Keys.GetTimes() - 1
			if ln >= buf.Lines.Len {
				ln = buf.Lines.Len - 1
			}
			buf.Cursor.Line = ln

			_, h := e.View.Size()
			if buf.Cursor.Line > buf.ScrollOffset+h-3 {
				buf.ScrollOffset = buf.Cursor.Line - h/2
			}

			if buf.Cursor.Line <= buf.ScrollOffset+h+3 {
				buf.ScrollOffset = buf.Cursor.Line - h/2
			}

			e.Keys.resetState()
		}
	})
}

func CmdGotoLineEnd(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) > 0 {
			buf.Cursor.Char = len(line.Value) - 1
		} else {
			buf.Cursor.Char = 0
		}
	})
}

func CmdForwardWord(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			_, h := e.View.Size()
			if buf.Cursor.Line-buf.ScrollOffset > h-3 {
				CmdScrollDown(e)
			}
		}()

		cls := CursorChClass(buf)
		CursorInc(buf)

		// return on line change
		if line != CursorLine(buf) {
			return
		}

		if cls != chWhitespace {
			for CursorChClass(buf) == cls {
				if !CursorInc(buf) {
					return
				}
			}
		}

		// skip whitespace
		line = CursorLine(buf)
		for CursorChClass(buf) == chWhitespace {
			if !CursorInc(buf) {
				return
			}
			if line != CursorLine(buf) {
				return
			}
		}
	})
}

func CmdBackwardWord(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			if buf.Cursor.Line < buf.ScrollOffset+3 {
				CmdScrollUp(e)
			}
		}()

		cls := CursorChClass(buf)
		CursorDec(buf)

		// return on line change
		if line != CursorLine(buf) {
			return
		}

		if cls != chWhitespace && CursorChClass(buf) == cls {
			for {
				if buf.Cursor.Char == 0 {
					return
				}
				if CursorChClass(buf) != cls {
					CursorInc(buf)
					return
				}

				if !CursorDec(buf) {
					return
				}
			}
		}

		// skip !=cls and whitespace
		for CursorChClass(buf) == chWhitespace {
			if !CursorDec(buf) {
				return
			}
		}

		cls = CursorChClass(buf)
		for {
			if buf.Cursor.Char == 0 {
				return
			}
			if CursorChClass(buf) == cls {
				if !CursorDec(buf) {
					return
				}
				continue
			}
			CursorInc(buf)
			break
		}
	})
}

func CmdReplaceChar(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		c := []rune(ch)
		line.Value[buf.Cursor.Char] = c[0]
	})
}

func CmdForwardToChar(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if line.Value.IsEmpty() {
			return
		}
		for i := buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ch) {
				buf.Cursor.Char = i
				buf.Cursor.PreserveCharPosition = i
				break
			}
		}
	})
}

func CmdForwardBeforeChar(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if line.Value.IsEmpty() {
			return
		}
		for i := buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ch) {
				buf.Cursor.Char = i - 1
				buf.Cursor.PreserveCharPosition = i - 1
				break
			}
		}
	})
}

func CmdBackwardChar(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Line == 0 && buf.Cursor.Char == 0 {
			return
		}

		if len(line.Value) == 0 {
			if buf.Lines.Len > 1 {
				buf.Lines.Remove(line)
			}
			CmdCursorLineUp(e)
			CmdAppendLine(e)
			return
		}

		if buf.Cursor.Char == 0 {
			CmdCursorLineUp(e)
			line = CursorLine(buf)
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
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
	CmdIndent(e)
}

func CmdLineOpenAbove(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
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
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if line == nil {
			return
		}

		CmdYank(e)

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

func CmdDeleteWord(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		_, end := TextObjectWord(buf, false)
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   Cursor{Line: buf.Cursor.Line, Char: end},
		}
		CmdSelectinDelete(e)
	})
}

func CmdChangeWord(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		_, end := TextObjectWord(buf, false)
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   Cursor{Line: buf.Cursor.Line, Char: end},
		}
		CmdSelectinDelete(e)
		CmdInsertMode(e)
	})
}

func CmdChangeWORD(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		start, end := TextObjectWord(buf, true)
		buf.Cursor.Char = start
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   Cursor{Line: buf.Cursor.Line, Char: end},
		}
		CmdSelectinDelete(e)
		CmdInsertMode(e)
	})
}

func CmdChangeTo(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardToChar)(e, ch)
		CmdSelectionChange(e)
	})
}

func CmdChangeBefore(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardBeforeChar)(e, ch)
		CmdSelectionChange(e)
	})
}

func CmdChangeEndOfLine(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		buf.Selection.End.Char = len(line.Value) - 1
		CmdSelectionChange(e)
	})
}

func CmdChangeLine(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		CmdInsertModeAfter(e)
		line.Value = nil
	})
}

func CmdDeleteTo(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardToChar)(e, ch)
		CmdSelectinDelete(e)
	})
}

func CmdDeleteBefore(e *Editor, ch string) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardBeforeChar)(e, ch)
		CmdSelectinDelete(e)
	})
}

func CmdSelectionChange(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		CmdSelectinDelete(e)
		CmdInsertMode(e)
	})
}

func CmdSelectinDelete(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			CmdNormalMode(e)
		}()
		if buf.Selection == nil {
			return
		}
		sel := SelectionNormalize(buf.Selection)

		yankSave(e, buf, line)

		lineStart := CursorLineByNum(buf, sel.Start.Line)
		lineEnd := CursorLineByNum(buf, sel.End.Line)

		if sel.Start.Line == sel.End.Line {
			if sel.End.Char < len(lineStart.Value) {
				lineStart.Value = append(lineStart.Value[:sel.Start.Char], lineStart.Value[sel.End.Char+1:]...)
			} else {
				lineStart.Value = lineStart.Value[:sel.Start.Char]
			}

			if len(lineStart.Value) == 0 {
				buf.Lines.Remove(lineStart)
				CmdCursorBeginningOfTheLine(e)
				return
			}

			cursorGotoChar(buf, sel.Start.Char)
		} else {
			// delete all lines between start and end line
			for lineStart.Next() != lineEnd {
				buf.Lines.Remove(lineStart.Next())
			}

			lineStart.Value = lineStart.Value[:sel.Start.Char]

			if sel.End.Char+1 <= len(lineEnd.Value) {
				lineEnd.Value = lineEnd.Value[sel.End.Char+1:]
			}

			if len(lineEnd.Value) == 0 {
				buf.Lines.Remove(lineEnd)
			}

			lineJoinNext(buf, lineStart)

			buf.Cursor.Line = sel.Start.Line
			if lineStart != nil && sel.Start.Char < len(lineStart.Value) {
				cursorGotoChar(buf, sel.Start.Char)
			} else {
				CmdGotoLineEnd(e)
			}
		}
	})
}

func CmdSaveFile(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		err := buf.Save()
		var msg string
		if err == nil {
			msg = fmt.Sprintf("Saved file %s. Lines: %d.", buf.FilePath, buf.Lines.Len)
		} else {
			msg = err.Error()
		}

		e.LogMessage(msg)
		e.EchoMessage(msg)
	})
}

func CmdWindowVSplit(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
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

func CmdWindowToggleLayout(e *Editor) {
	if e.Layout == LayoutHorizontal {
		e.Layout = LayoutVertical
	} else {
		e.Layout = LayoutHorizontal
	}
}

func CmdWindowClose(e *Editor) {
	if len(e.Windows) == 1 {
		return
	}

	curWin := e.activeWindow
	for i, w := range e.Windows {
		if w == curWin {
			e.Windows = append(e.Windows[:i], e.Windows[i+1:]...)
			e.activeWindow = e.Windows[0]
		}
	}
}

func CmdExit(e *Editor) {
	e.ExitCh <- 1
}

func CmdYank(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			if buf.Selection != nil {
				buf.Cursor = buf.Selection.Start
			}
			CmdNormalMode(e)
		}()
		yankSave(e, buf, line)
	})
}

func CmdYankPut(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if e.Yanks.Len == 0 {
			return
		}

		CmdCursorRight(e)
		v := e.Yanks.Last()

		if v.Value.isLine {
			CmdGotoLineEnd(e)
			CmdCursorRight(e)
			CmdNewLine(e)
			CmdEnsureCursorVisible(e)
			defer CmdCursorBeginningOfTheLine(e)
		}

		yankPut(e, buf)
	})
}

func CmdYankPutBefore(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if e.Yanks.Len == 0 {
			return
		}

		v := e.Yanks.Last()
		if v.Value.isLine {
			CmdLineOpenAbove(e)
			CmdNormalMode(e)
			yankPut(e, buf)
			CmdCursorBeginningOfTheLine(e)
		} else {
			yankPut(e, buf)
		}
	})
}

func CmdKillBuffer(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		if len(e.Buffers) == 0 {
			return
		}

		// creates [No Name] buffer
		defer e.ActiveBuffer()

		// remove from buffers list
		for i, b := range e.Buffers {
			if b == buf {
				e.Buffers = append(e.Buffers[:i], e.Buffers[i+1:]...)
				if len(e.Buffers) > 0 {
					idx := i - 1
					if idx < 0 {
						idx = 0
					}
					e.activeWindow.Buffer = e.Buffers[idx]
				}
			}
		}

		// cleanup all nodes
		{
			l := buf.Lines.First()
			for l != nil {
				next := l.Next()
				l.Value = nil
				buf.Lines.Remove(l)
				l = next
			}
		}

	})
}

func CmdEnsureCursorVisible(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		_, h := e.View.Size()
		if buf.Cursor.Line > buf.ScrollOffset+h-3 {
			buf.ScrollOffset = buf.Cursor.Line - h + 3
		}

		if buf.Cursor.Line < buf.ScrollOffset+3 {
			buf.ScrollOffset = buf.Cursor.Line - 3
		}
	})
}

func CmdChangeInsideBlock(e *Editor, ch string) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		switch ch {
		case "w":
			CmdChangeWORD(e)
		case "(", "[", "{", "'", "\"":
			found, sel, cur := TextObjectBlock(buf, rune(ch[0]), false)
			if !found {
				return
			}
			buf.Selection = sel
			buf.Cursor = cur
			CmdSelectionChange(e)
		}
	})
}

func CmdUndo(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		buf.UndoRedo.Undo()
	})
}

func CmdRedo(e *Editor) {
	Do(e, func(buf *Buffer, _ *Element[Line]) {
		buf.UndoRedo.Redo()
	})
}

func WithSelection(fn func(*Editor)) func(*Editor) {
	return func(e *Editor) {
		fn(e)
		buf := e.ActiveBuffer()
		buf.Selection.End = buf.Cursor

		if buf.Mode() == MODE_VISUAL_LINE {
			if buf.Selection.Start.Line > buf.Selection.End.Line {} else {
				lineEnd := CursorLineByNum(buf, buf.Selection.End.Line)
				buf.Selection.Start.Char = 0
				buf.Selection.End.Char = len(lineEnd.Value) - 1
			}
		}
	}
}

func WithSelectionToChar(fn func(*Editor, string)) func(*Editor, string) {
	return func(e *Editor, ch string) {
		fn(e, ch)
		buf := e.ActiveBuffer()
		buf.Selection.End = buf.Cursor
	}
}
