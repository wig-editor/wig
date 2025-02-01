package mcwig

import (
	"fmt"
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

func CmdJoinNextLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	lineJoinNext(ctx.Buf, CursorLine(ctx.Buf))
}

func CmdEnterInsertMode(ctx Context) {
	line := CursorLine(ctx.Buf)
	if line == nil {
		return
	}

	if !ctx.Buf.TxStart() {
		ctx.Editor.LogMessage("should not happen")
	}

	if len(line.Value) == 0 {
		ctx.Buf.Cursor.Char++
	}

	ctx.Buf.SetMode(MODE_INSERT)
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
	SelectionStart(ctx.Buf)
	ctx.Buf.SetMode(MODE_VISUAL)
}
func CmdNormalMode(ctx Context) {
	if ctx.Buf.Mode() == MODE_INSERT {
		line := CursorLine(ctx.Buf)
		CmdCursorLeft(ctx)
		if ctx.Buf.Cursor.Char >= len(line.Value) {
			CmdGotoLineEnd(ctx)
		}
	}
	ctx.Buf.SetMode(MODE_NORMAL)
	ctx.Buf.Selection = nil
}

func CmdVisualLineMode(ctx Context) {
	line := CursorLine(ctx.Buf)
	SelectionStart(ctx.Buf)
	ctx.Buf.Selection.Start.Char = 0
	ctx.Buf.Selection.End.Char = len(line.Value) - 1
	ctx.Buf.SetMode(MODE_VISUAL_LINE)
}

func CmdInsertModeAfter(ctx Context) {
	ctx.Buf.Cursor.Char++
	CmdEnterInsertMode(ctx)
}

func CmdReplaceChar(ctx Context) func(Context) {
	return func(ctx Context) {
		c := []rune(ctx.Char)
		line := CursorLine(ctx.Buf)
		line.Value[ctx.Buf.Cursor.Char] = c[0]
	}
}

func CmdDeleteCharForward(ctx Context) {
	line := CursorLine(ctx.Buf)
	if len(line.Value) == 0 {
		CmdGotoLineEnd(ctx)
		lineJoinNext(ctx.Buf, line)
		CmdCursorBeginningOfTheLine(ctx)
		return
	}
	line.Value = append(line.Value[:ctx.Buf.Cursor.Char], line.Value[ctx.Buf.Cursor.Char+1:]...)
	if ctx.Buf.Cursor.Char >= len(line.Value) {
		CmdCursorLeft(ctx)
	}
}

func CmdDeleteCharBackward(ctx Context) {
	if ctx.Buf.Cursor.Line == 0 && ctx.Buf.Cursor.Char == 0 {
		return
	}

	line := CursorLine(ctx.Buf)
	if len(line.Value) == 0 {
		if ctx.Buf.Lines.Len > 1 {
			ctx.Buf.Lines.Remove(line)
		}
		CmdCursorLineUp(ctx)
		CmdAppendLine(ctx)
		return
	}

	if ctx.Buf.Cursor.Char == 0 {
		CmdCursorLineUp(ctx)
		line = CursorLine(ctx.Buf)
		pos := len(line.Value)
		lineJoinNext(ctx.Buf, line)
		cursorGotoChar(ctx.Buf, pos)
		return
	}

	if ctx.Buf.Cursor.Char >= len(line.Value) && len(line.Value) > 0 {
		line.Value = line.Value[:len(line.Value)-1]
		CmdCursorLeft(ctx)
		return
	}

	CmdCursorLeft(ctx)
	CmdDeleteCharForward(ctx)
}

func CmdAppendLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	CmdInsertModeAfter(ctx)
}

func CmdNewLine(ctx Context) {
	line := CursorLine(ctx.Buf)
	// EOL
	if (ctx.Buf.Cursor.Char) >= len(line.Value) {
		ctx.Buf.Lines.insertValueAfter(Line{}, line)
		ctx.Buf.Cursor.Line++
		ctx.Buf.Cursor.Char = 1
		ctx.Buf.Cursor.PreserveCharPosition = 0
		return
	}

	// split line
	tmpData := make([]rune, len(line.Value[ctx.Buf.Cursor.Char:]))
	copy(tmpData, line.Value[ctx.Buf.Cursor.Char:])
	line.Value = line.Value[:ctx.Buf.Cursor.Char]
	ctx.Buf.Lines.insertValueAfter(tmpData, line)
	CmdCursorLineDown(ctx)
	CmdCursorBeginningOfTheLine(ctx)
}

func CmdLineOpenBelow(ctx Context) {
	CmdAppendLine(ctx)
	CmdNewLine(ctx)
	CmdInsertModeAfter(ctx)
	CmdIndent(ctx)
}

func CmdLineOpenAbove(ctx Context) {
	if ctx.Buf.Cursor.Line == 0 {
		ctx.Buf.Lines.PushFront(Line{})
		CmdCursorBeginningOfTheLine(ctx)
		CmdInsertModeAfter(ctx)
		return
	}
	CmdCursorLineUp(ctx)
	CmdLineOpenBelow(ctx)
}

func CmdDeleteLine(ctx Context) {
	CmdVisualLineMode(ctx)
	ctx.Buf.Selection.End.Line = ctx.Buf.Selection.Start.Line + int(ctx.Count)
	ctx.Buf.Selection.End.Char = len(CursorLineByNum(ctx.Buf, ctx.Buf.Selection.End.Line).Value)
	CmdSelectinDelete(ctx)
	CmdNormalMode(ctx)
}

func CmdDeleteWord(ctx Context) {
	_, end := TextObjectWord(ctx.Buf, false)
	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: end},
	}
	CmdSelectinDelete(ctx)
}

func CmdChangeWord(ctx Context) {
	_, end := TextObjectWord(ctx.Buf, false)
	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: end},
	}
	CmdSelectinDelete(ctx)
	CmdEnterInsertMode(ctx)
}

func CmdChangeWORD(ctx Context) {
	start, end := TextObjectWord(ctx.Buf, true)
	ctx.Buf.Cursor.Char = start
	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: end},
	}
	CmdSelectinDelete(ctx)
	CmdEnterInsertMode(ctx)
}

func CmdChangeTo(_ Context) func(Context) {
	return func(ctx Context) {
		SelectionStart(ctx.Buf)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf)
		CmdSelectionChange(ctx)
	}
}

func CmdChangeBefore(_ Context) func(Context) {
	return func(ctx Context) {
		SelectionStart(ctx.Buf)
		CmdForwardBeforeChar(ctx)(ctx)
		SelectionStop(ctx.Buf)
		CmdSelectionChange(ctx)
	}
}

func CmdChangeEndOfLine(ctx Context) {
	SelectionStart(ctx.Buf)
	CmdGotoLineEnd(ctx)
	SelectionStop(ctx.Buf)
	CmdSelectionChange(ctx)
}

func CmdChangeLine(ctx Context) {
	CmdInsertModeAfter(ctx)
	line := CursorLine(ctx.Buf)
	line.Value = nil
}

func CmdDeleteTo(_ Context) func(Context) {
	return func(ctx Context) {
		SelectionStart(ctx.Buf)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf)
		CmdSelectinDelete(ctx)
	}
}

func CmdDeleteBefore(ctx Context) {
	SelectionStart(ctx.Buf)
	CmdForwardBeforeChar(ctx)(ctx)
	CmdSelectinDelete(ctx)
}

func CmdSelectionChange(ctx Context) {
	CmdSelectinDelete(ctx)
	CmdEnterInsertMode(ctx)
}

func CmdToggleComment(ctx Context) {
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
	line := CursorLine(ctx.Buf)
	cmAppend(line)
}

func CmdSelectinDelete(ctx Context) {
	defer func() {
		ctx.Buf.Selection = nil
	}()
	if ctx.Buf.Selection == nil {
		return
	}

	sel := SelectionNormalize(ctx.Buf.Selection)

	yankSave(ctx)

	lineStart := CursorLineByNum(ctx.Buf, sel.Start.Line)
	lineEnd := CursorLineByNum(ctx.Buf, sel.End.Line)

	if sel.Start.Line == sel.End.Line {
		if len(lineStart.Value) == 0 || ctx.Buf.Mode() == MODE_VISUAL_LINE {
			ctx.Buf.Lines.Remove(lineStart)
			CmdCursorBeginningOfTheLine(ctx)
			return
		}

		if sel.End.Char < len(lineStart.Value) {
			lineStart.Value = append(lineStart.Value[:sel.Start.Char], lineStart.Value[sel.End.Char+1:]...)
		} else {
			lineStart.Value = lineStart.Value[:sel.Start.Char]
		}

		cursorGotoChar(ctx.Buf, sel.Start.Char)
	} else {
		// delete all lines between start and end line
		for lineStart.Next() != lineEnd {
			ctx.Buf.Lines.Remove(lineStart.Next())
		}

		lineStart.Value = lineStart.Value[:sel.Start.Char]

		if sel.End.Char+1 <= len(lineEnd.Value) {
			lineEnd.Value = lineEnd.Value[sel.End.Char+1:]
		}

		if len(lineEnd.Value) == 0 {
			ctx.Buf.Lines.Remove(lineEnd)
		}

		lineJoinNext(ctx.Buf, lineStart)

		ctx.Buf.Cursor.Line = sel.Start.Line
		if lineStart != nil && sel.Start.Char < len(lineStart.Value) {
			cursorGotoChar(ctx.Buf, sel.Start.Char)
		} else {
			CmdGotoLineEnd(ctx)
		}
	}
}

func CmdSaveFile(ctx Context) {
	err := ctx.Buf.Save()
	var msg string
	if err == nil {
		msg = fmt.Sprintf("Saved file %s. Lines: %d.", ctx.Buf.FilePath, ctx.Buf.Lines.Len)
	} else {
		msg = err.Error()
	}

	ctx.Editor.LogMessage(msg)
	ctx.Editor.EchoMessage(msg)
}

func CmdKillBuffer(ctx Context) {
	buffers := ctx.Editor.Buffers
	if len(buffers) == 0 {
		return
	}

	// creates [No Name] buffer
	defer ctx.Editor.ActiveBuffer()

	// remove from buffers list
	// ands moves to the next buffer
	for i, b := range buffers {
		if b == ctx.Buf {
			buffers = append(buffers[:i], buffers[i+1:]...)
			if len(buffers) > 0 {
				idx := i - 1
				if idx < 0 {
					idx = 0
				}
				ctx.Editor.ActiveWindow().VisitBuffer(buffers[idx])
			}
		}
	}

	// cleanup all nodes
	{
		l := ctx.Buf.Lines.First()
		for l != nil {
			next := l.Next()
			l.Value = nil
			ctx.Buf.Lines.Remove(l)
			l = next
		}
	}

	ctx.Editor.Lsp.DidClose(ctx.Buf)
}

func CmdIndentOrComplete(ctx Context) {
	ctx.Editor.Lsp.Completion(ctx.Buf)
}

func CmdChangeInsideBlock(ctx Context) {
	switch ctx.Char {
	case "w":
		CmdChangeWORD(ctx)
	case "(", "[", "{", "'", "\"":
		found, sel, cur := TextObjectBlock(ctx.Buf, rune(ctx.Char[0]), false) // TODO: handle unicode
		if !found {
			return
		}
		ctx.Buf.Selection = sel
		ctx.Buf.Cursor = cur
		CmdSelectionChange(ctx)
	}
}

func CmdUndo(ctx Context) {
	ctx.Buf.UndoRedo.Undo()
}

func CmdRedo(ctx Context) {
	ctx.Buf.UndoRedo.Redo()
}

func CmdExit(ctx Context) {
	ctx.Editor.ExitCh <- 1
}
