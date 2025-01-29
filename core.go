package mcwig

import (
	"fmt"
	"strings"
	"unicode"
)

const minVisibleLines = 6

func lineJoinNext(buf *Buffer, line *Element[Line]) {
	next := line.Next()
	if next == nil {
		return
	}
	line.Value = append(line.Value, next.Value...)
	buf.Lines.Remove(next)
}

func Do(ctx Context, fn func(buf *Buffer, line *Element[Line])) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	fn(ctx.Buf, CursorLine(ctx.Buf))
}

func CmdJoinNextLine(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		CmdGotoLineEnd(ctx)
		lineJoinNext(buf, line)
	})
}

func CmdScrollUp(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if buf.ScrollOffset > 0 {
			buf.ScrollOffset--

			_, h := ctx.Editor.View.Size()
			if buf.Cursor.Line > buf.ScrollOffset+h-minVisibleLines {
				CmdCursorLineUp(ctx)
			}
		}
	})
}

func CmdScrollDown(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if buf.ScrollOffset < buf.Lines.Len-minVisibleLines {
			buf.ScrollOffset++

			if buf.Cursor.Line <= buf.ScrollOffset+minVisibleLines {
				CmdCursorLineDown(ctx)
			}
		}
	})
}

func CmdCursorLeft(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		if buf.Cursor.Char > 0 {
			buf.Cursor.Char--
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	})
}

func CmdCursorRight(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Char < len(line.Value) {
			buf.Cursor.Char++
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	})
}

func CmdCursorLineUp(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf == nil {
		return
	}
	if buf.Cursor.Line > 0 {
		buf.Cursor.Line--
		restoreCharPosition(buf)

		if buf.Cursor.Line < buf.ScrollOffset+minVisibleLines {
			CmdScrollUp(ctx)
		}
	}
}

func CmdCursorLineDown(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf == nil {
		return
	}
	if buf.Cursor.Line < buf.Lines.Len-1 {
		buf.Cursor.Line++
		restoreCharPosition(buf)

		_, h := ctx.Editor.View.Size()
		if buf.Cursor.Line-buf.ScrollOffset > h-minVisibleLines {
			CmdScrollDown(ctx)
		}
	}
}

func CmdCursorBeginningOfTheLine(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Char = 0
		buf.Cursor.PreserveCharPosition = 0
	})
}

func CmdCursorFirstNonBlank(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		CmdCursorBeginningOfTheLine(ctx)
		if len(line.Value) == 0 {
			return
		}
		for _, c := range line.Value {
			if unicode.IsSpace(c) {
				CmdCursorRight(ctx)
			} else {
				break
			}
		}
	})
}

func CmdEnterInsertMode(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf == nil {
		return
	}

	line := CursorLine(buf)
	if line == nil {
		return
	}

	if !buf.TxStart() {
		ctx.Editor.LogMessage("should not happen")
	}

	if len(line.Value) == 0 {
		buf.Cursor.Char++
	}

	buf.SetMode(MODE_INSERT)
}

func CmdExitInsertMode(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf == nil {
		return
	}

	defer func() {
		buf.SetMode(MODE_NORMAL)
		buf.Selection = nil
	}()

	line := CursorLine(buf)
	if line == nil {
		return
	}

	CmdCursorLeft(ctx)
	if buf.Cursor.Char >= len(line.Value) {
		CmdGotoLineEnd(ctx)
	}

	buf.TxEnd()

	// TODO: this is ugly
	if buf.Highlighter != nil {
		buf.Highlighter.Build()
	}
}

func CmdVisualMode(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		SelectionStart(buf)
		buf.SetMode(MODE_VISUAL)
	})
}
func CmdNormalMode(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if buf.Mode() == MODE_INSERT {
			CmdCursorLeft(ctx)
			if buf.Cursor.Char >= len(line.Value) {
				CmdGotoLineEnd(ctx)
			}
		}
		buf.SetMode(MODE_NORMAL)
		buf.Selection = nil
	})
}

func CmdVisualLineMode(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		buf.Selection.Start.Char = 0
		buf.Selection.End.Char = len(line.Value) - 1
		buf.SetMode(MODE_VISUAL_LINE)
	})
}

func CmdInsertModeAfter(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		buf.Cursor.Char++
		CmdEnterInsertMode(ctx)
	})
}

func CmdGotoLine0(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		buf.Cursor.Line = 0
		buf.ScrollOffset = 0
		restoreCharPosition(buf)
		ctx.Editor.ActiveWindow().Jumps.Push(buf)

		if ctx.Editor.Keys.GetTimes() > 1 {
			ln := ctx.Editor.Keys.GetTimes() - 1
			if ln >= buf.Lines.Len {
				ln = buf.Lines.Len - 1
			}
			buf.Cursor.Line = ln

			_, h := ctx.Editor.View.Size()
			if buf.Cursor.Line > buf.ScrollOffset+h-minVisibleLines {
				buf.ScrollOffset = buf.Cursor.Line - h/2
			}

			if buf.Cursor.Line <= buf.ScrollOffset+h+minVisibleLines {
				buf.ScrollOffset = buf.Cursor.Line - h/2
			}

			ctx.Editor.Keys.resetState()
		}
	})
}

func CmdGotoLineEnd(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if line == nil {
			return
		}
		if len(line.Value) > 0 {
			buf.Cursor.Char = len(line.Value) - 1
		} else {
			buf.Cursor.Char = 0
		}
		ctx.Editor.ActiveWindow().Jumps.Push(buf)
	})
}

func CmdForwardWord(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			_, h := ctx.Editor.View.Size()
			if buf.Cursor.Line-buf.ScrollOffset > h-minVisibleLines {
				CmdScrollDown(ctx)
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

func CmdBackwardWord(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			if buf.Cursor.Line < buf.ScrollOffset+minVisibleLines {
				CmdScrollUp(ctx)
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

func CmdReplaceChar(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		c := []rune(ctx.Char)
		line.Value[buf.Cursor.Char] = c[0]
	})
}

func CmdForwardToChar(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if line.Value.IsEmpty() {
			return
		}
		for i := buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ctx.Char) {
				buf.Cursor.Char = i
				buf.Cursor.PreserveCharPosition = i
				break
			}
		}
	})
}

func CmdForwardBeforeChar(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if line.Value.IsEmpty() {
			return
		}
		for i := buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ctx.Char) {
				buf.Cursor.Char = i - 1
				buf.Cursor.PreserveCharPosition = i - 1
				break
			}
		}
	})
}

func CmdBackwardChar(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			return
		}

		for i := buf.Cursor.Char - 1; i >= 0; i-- {
			if string(line.Value[i]) == ctx.Char {
				buf.Cursor.Char = i
				buf.Cursor.PreserveCharPosition = i
				break
			}
		}
	})
}

func CmdDeleteCharForward(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if len(line.Value) == 0 {
			CmdGotoLineEnd(ctx)
			lineJoinNext(buf, line)
			CmdCursorBeginningOfTheLine(ctx)
			return
		}
		line.Value = append(line.Value[:buf.Cursor.Char], line.Value[buf.Cursor.Char+1:]...)
		if buf.Cursor.Char >= len(line.Value) {
			CmdCursorLeft(ctx)
		}
	})
}

func CmdDeleteCharBackward(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if buf.Cursor.Line == 0 && buf.Cursor.Char == 0 {
			return
		}

		if len(line.Value) == 0 {
			if buf.Lines.Len > 1 {
				buf.Lines.Remove(line)
			}
			CmdCursorLineUp(ctx)
			CmdAppendLine(ctx)
			return
		}

		if buf.Cursor.Char == 0 {
			CmdCursorLineUp(ctx)
			line = CursorLine(buf)
			pos := len(line.Value)
			lineJoinNext(buf, line)
			cursorGotoChar(buf, pos)
			return
		}

		if buf.Cursor.Char >= len(line.Value) && len(line.Value) > 0 {
			line.Value = line.Value[:len(line.Value)-1]
			CmdCursorLeft(ctx)
			return
		}

		CmdCursorLeft(ctx)
		CmdDeleteCharForward(ctx)
	})
}

func CmdAppendLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	CmdInsertModeAfter(ctx)
}

func CmdNewLine(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
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
		CmdCursorLineDown(ctx)
		CmdCursorBeginningOfTheLine(ctx)
	})
}

func CmdLineOpenBelow(ctx Context) {
	CmdAppendLine(ctx)
	CmdNewLine(ctx)
	CmdInsertModeAfter(ctx)
	CmdIndent(ctx)
}

func CmdLineOpenAbove(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		if buf.Cursor.Line == 0 {
			buf.Lines.PushFront(Line{})
			CmdCursorBeginningOfTheLine(ctx)
			CmdInsertModeAfter(ctx)
			return
		}
		CmdCursorLineUp(ctx)
		CmdLineOpenBelow(ctx)
	})
}

func CmdDeleteLine(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		CmdVisualLineMode(ctx)
		CmdSelectinDelete(ctx)
		CmdNormalMode(ctx)
	})
}

func CmdDeleteWord(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		_, end := TextObjectWord(buf, false)
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   Cursor{Line: buf.Cursor.Line, Char: end},
		}
		CmdSelectinDelete(ctx)
	})
}

func CmdChangeWord(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		_, end := TextObjectWord(buf, false)
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   Cursor{Line: buf.Cursor.Line, Char: end},
		}
		CmdSelectinDelete(ctx)
		CmdEnterInsertMode(ctx)
	})
}

func CmdChangeWORD(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		start, end := TextObjectWord(buf, true)
		buf.Cursor.Char = start
		buf.Selection = &Selection{
			Start: buf.Cursor,
			End:   Cursor{Line: buf.Cursor.Line, Char: end},
		}
		CmdSelectinDelete(ctx)
		CmdEnterInsertMode(ctx)
	})
}

func CmdChangeTo(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardToChar)(ctx)
		CmdSelectionChange(ctx)
	})
}

func CmdChangeBefore(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardBeforeChar)(ctx)
		CmdSelectionChange(ctx)
	})
}

func CmdChangeEndOfLine(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		buf.Selection.End.Char = len(line.Value) - 1
		CmdSelectionChange(ctx)
	})
}

func CmdChangeLine(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		CmdInsertModeAfter(ctx)
		line.Value = nil
	})
}

func CmdDeleteTo(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardToChar)(ctx)
		CmdSelectinDelete(ctx)
	})
}

func CmdDeleteBefore(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		SelectionStart(buf)
		WithSelectionToChar(CmdForwardBeforeChar)(ctx)
		CmdSelectinDelete(ctx)
	})
}

func CmdSelectionChange(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		CmdSelectinDelete(ctx)
		CmdEnterInsertMode(ctx)
	})
}

func CmdToggleComment(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		comment := []rune("// ")
		cmAppend := func(line *Element[Line]) {
			idx := 0
			for i, c := range line.Value {
				if !unicode.IsSpace(c) {
					idx = i
					break
				}
			}
			tmpData := make([]rune, 0, len(line.Value)+len(comment))
			tmpData = append(tmpData, line.Value[:idx]...)
			tmpData = append(tmpData, comment...)
			tmpData = append(tmpData, line.Value[idx:]...)
			line.Value = tmpData

		}
		cmAppend(line)
	})
}

func CmdSelectinDelete(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			buf.Selection = nil
		}()
		if buf.Selection == nil {
			return
		}

		sel := SelectionNormalize(buf.Selection)

		yankSave(ctx.Editor, buf, line)

		lineStart := CursorLineByNum(buf, sel.Start.Line)
		lineEnd := CursorLineByNum(buf, sel.End.Line)

		if sel.Start.Line == sel.End.Line {
			if len(lineStart.Value) == 0 || buf.Mode() == MODE_VISUAL_LINE {
				buf.Lines.Remove(lineStart)
				CmdCursorBeginningOfTheLine(ctx)
				return
			}

			if sel.End.Char < len(lineStart.Value) {
				lineStart.Value = append(lineStart.Value[:sel.Start.Char], lineStart.Value[sel.End.Char+1:]...)
			} else {
				lineStart.Value = lineStart.Value[:sel.Start.Char]
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
				CmdGotoLineEnd(ctx)
			}
		}
	})
}

func CmdSaveFile(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		err := buf.Save()
		var msg string
		if err == nil {
			msg = fmt.Sprintf("Saved file %s. Lines: %d.", buf.FilePath, buf.Lines.Len)
		} else {
			msg = err.Error()
		}

		ctx.Editor.LogMessage(msg)
		ctx.Editor.EchoMessage(msg)
	})
}

func CmdWindowVSplit(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		nwin := CreateWindow()
		nwin.VisitBuffer(buf)
		ctx.Editor.Windows = append(ctx.Editor.Windows, nwin)
	})
}

func CmdWindowNext(ctx Context) {
	curWin := ctx.Editor.activeWindow
	idx := 0
	for i, w := range ctx.Editor.Windows {
		if w == curWin {
			idx = i + 1
			break
		}
	}

	if idx >= len(ctx.Editor.Windows) {
		idx = 0
	}

	ctx.Editor.activeWindow = ctx.Editor.Windows[idx]
}

func CmdWindowToggleLayout(ctx Context) {
	if ctx.Editor.Layout == LayoutHorizontal {
		ctx.Editor.Layout = LayoutVertical
	} else {
		ctx.Editor.Layout = LayoutHorizontal
	}
}

func CmdWindowClose(ctx Context) {
	if len(e.Windows) == 1 {
		return
	}

	curWin := ctx.Editor.activeWindow
	for i, w := range ctx.Editor.Windows {
		if w == curWin {
			ctx.Editor.Windows = append(ctx.Editor.Windows[:i], ctx.Editor.Windows[i+1:]...)
			ctx.Editor.activeWindow = ctx.Editor.Windows[0]
		}
	}
}

func CmdExit(ctx Context) {
	ctx.Editor.ExitCh <- 1
}

func CmdYank(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		defer func() {
			if buf.Selection != nil {
				buf.Cursor = buf.Selection.Start
			}
			CmdExitInsertMode(ctx)
		}()
		yankSave(ctx.Editor, buf, line)
	})
}

func CmdYankPut(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if ctx.Editor.Yanks.Len == 0 {
			return
		}

		CmdCursorRight(ctx)
		v := ctx.Editor.Yanks.Last()

		if v.Value.isLine {
			CmdGotoLineEnd(ctx)
			CmdCursorRight(ctx)
			CmdNewLine(ctx)
			CmdEnsureCursorVisible(ctx)
			defer CmdCursorBeginningOfTheLine(ctx)
		}

		yankPut(ctx.Editor, buf)
	})
}

func CmdYankPutBefore(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if ctx.Editor.Yanks.Len == 0 {
			return
		}

		v := ctx.Editor.Yanks.Last()
		if v.Value.isLine {
			CmdLineOpenAbove(ctx)
			CmdExitInsertMode(ctx)
			yankPut(ctx.Editor, buf)
			CmdCursorBeginningOfTheLine(ctx)
		} else {
			yankPut(ctx.Editor, buf)
		}
	})
}

func CmdKillBuffer(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		if len(e.Buffers) == 0 {
			return
		}

		// creates [No Name] buffer
		defer ctx.Editor.ActiveBuffer()

		// remove from buffers list
		// ands moves to the next buffer
		for i, b := range ctx.Editor.Buffers {
			if b == buf {
				e.Buffers = append(e.Buffers[:i], ctx.Editor.Buffers[i+1:]...)
				if len(e.Buffers) > 0 {
					idx := i - 1
					if idx < 0 {
						idx = 0
					}
					e.ActiveWindow().VisitBuffer(e.Buffers[idx])
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

		e.Lsp.DidClose(buf)

	})
}

func CmdIndentOrComplete(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		e.Lsp.Completion(buf)
	})
}

func CmdEnsureCursorVisible(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		_, h := ctx.Editor.View.Size()
		if buf.Cursor.Line > buf.ScrollOffset+h-minVisibleLines {
			buf.ScrollOffset = buf.Cursor.Line - h + minVisibleLines
		}

		if buf.Cursor.Line < buf.ScrollOffset+minVisibleLines {
			buf.ScrollOffset = buf.Cursor.Line - minVisibleLines
		}
	})
}

func CmdCursorCenter(ctx Context) {
	Do(ctx, func(buf *Buffer, line *Element[Line]) {
		_, h := ctx.Editor.View.Size()
		buf.ScrollOffset = buf.Cursor.Line - (h / 2) + minVisibleLines
	})
}

func CmdChangeInsideBlock(ctx Context) {
	Do(ctx, func(buf *Buffer, _ *Element[Line]) {
		switch ctx.Char {
		case "w":
			CmdChangeWORD(ctx)
		case "(", "[", "{", "'", "\"":
			found, sel, cur := TextObjectBlock(buf, rune(ctx.Char[0]), false) // TODO: handle unicode
			if !found {
				return
			}
			buf.Selection = sel
			buf.Cursor = cur
			CmdSelectionChange(ctx)
		}
	})
}

func CmdUndo(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf != nil {
		buf.UndoRedo.Undo()
	}
}

func CmdRedo(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf != nil {
		buf.UndoRedo.Redo()
	}
}

func CmdJumpBack(ctx Context) {
	e.ActiveWindow().Jumps.JumpBack()
	CmdCursorCenter(e)
}

func CmdJumpForward(ctx Context) {
	e.ActiveWindow().Jumps.JumpForward()
	CmdCursorCenter(e)
}

// Cycle between last two buffers in jump list
func CmdBufferCycle(ctx Context) {
	last := ctx.Editor.ActiveWindow().Jumps.List.Last()
	prev := last.Prev()

	if last == nil || prev == nil {
		return
	}

	if last.Value.FilePath == prev.Value.FilePath {
		return
	}

	var b *Buffer
	if last.Value.FilePath == ctx.Editor.ActiveWindow().Buffer().GetName() {
		b = ctx.Editor.BufferFindByFilePath(prev.Value.FilePath, false)
	} else {
		b = ctx.Editor.BufferFindByFilePath(last.Value.FilePath, false)
	}

	ctx.Editor.ActiveWindow().ShowBuffer(b)
}
