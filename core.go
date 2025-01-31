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

func CmdJoinNextLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	lineJoinNext(ctx.Buf, CursorLine(ctx.Buf))
}

func CmdScrollUp(ctx Context) {
	if ctx.Buf.ScrollOffset > 0 {
		ctx.Buf.ScrollOffset--

		_, h := ctx.Editor.View.Size()
		if ctx.Buf.Cursor.Line > ctx.Buf.ScrollOffset+h-minVisibleLines {
			CmdCursorLineUp(ctx)
		}
	}
}

func CmdScrollDown(ctx Context) {
	if ctx.Buf.ScrollOffset < ctx.Buf.Lines.Len-minVisibleLines {
		ctx.Buf.ScrollOffset++

		if ctx.Buf.Cursor.Line <= ctx.Buf.ScrollOffset+minVisibleLines {
			CmdCursorLineDown(ctx)
		}
	}
}

func CmdCursorLeft(ctx Context) {
	if ctx.Buf.Cursor.Char > 0 {
		ctx.Buf.Cursor.Char--
		ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
	}
}

func CmdCursorRight(ctx Context) {
	line := CursorLine(ctx.Buf)
	if ctx.Buf.Cursor.Char < len(line.Value) {
		ctx.Buf.Cursor.Char++
		ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
	}
}

func CmdCursorLineUp(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf == nil {
		return
	}
	if ctx.Buf.Cursor.Line > 0 {
		ctx.Buf.Cursor.Line--
		restoreCharPosition(buf)

		if ctx.Buf.Cursor.Line < ctx.Buf.ScrollOffset+minVisibleLines {
			CmdScrollUp(ctx)
		}
	}
}

func CmdCursorLineDown(ctx Context) {
	buf := ctx.Editor.ActiveBuffer()
	if buf == nil {
		return
	}
	if ctx.Buf.Cursor.Line < ctx.Buf.Lines.Len-1 {
		ctx.Buf.Cursor.Line++
		restoreCharPosition(buf)

		_, h := ctx.Editor.View.Size()
		if ctx.Buf.Cursor.Line-ctx.Buf.ScrollOffset > h-minVisibleLines {
			CmdScrollDown(ctx)
		}
	}
}

func CmdCursorBeginningOfTheLine(ctx Context) {
	ctx.Buf.Cursor.Char = 0
	ctx.Buf.Cursor.PreserveCharPosition = 0
}

func CmdCursorFirstNonBlank(ctx Context) {
	line := CursorLine(ctx.Buf)
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

func CmdGotoLine0(ctx Context) {
	ctx.Buf.Cursor.Line = 0
	ctx.Buf.ScrollOffset = 0
	restoreCharPosition(ctx.Buf)
	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf)
}

func CmdGotoLineEnd(ctx Context) {
	line := CursorLine(ctx.Buf)
	if len(line.Value) > 0 {
		ctx.Buf.Cursor.Char = len(line.Value) - 1
	} else {
		ctx.Buf.Cursor.Char = 0
	}
	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf)
}

func CmdForwardWord(ctx Context) {
	defer func() {
		CmdEnsureCursorVisible(ctx)
	}()

	line := CursorLine(ctx.Buf)

	cls := CursorChClass(ctx.Buf)
	CursorInc(ctx.Buf)

	// return on line change
	if line != CursorLine(ctx.Buf) {
		return
	}

	if cls != chWhitespace {
		for CursorChClass(ctx.Buf) == cls {
			if !CursorInc(ctx.Buf) {
				return
			}
		}
	}

	// skip whitespace
	line = CursorLine(ctx.Buf)
	for CursorChClass(ctx.Buf) == chWhitespace {
		if !CursorInc(ctx.Buf) {
			return
		}
		if line != CursorLine(ctx.Buf) {
			return
		}
	}
}

func CmdBackwardWord(ctx Context) {
	defer func() {
		CmdEnsureCursorVisible(ctx)
	}()

	line := CursorLine(ctx.Buf)
	cls := CursorChClass(ctx.Buf)
	CursorDec(ctx.Buf)

	// return on line change
	if line != CursorLine(ctx.Buf) {
		return
	}

	if cls != chWhitespace && CursorChClass(ctx.Buf) == cls {
		for {
			if ctx.Buf.Cursor.Char == 0 {
				return
			}
			if CursorChClass(ctx.Buf) != cls {
				CursorInc(ctx.Buf)
				return
			}

			if !CursorDec(ctx.Buf) {
				return
			}
		}
	}

	// skip !=cls and whitespace
	for CursorChClass(ctx.Buf) == chWhitespace {
		if !CursorDec(ctx.Buf) {
			return
		}
	}

	cls = CursorChClass(ctx.Buf)
	for {
		if ctx.Buf.Cursor.Char == 0 {
			return
		}
		if CursorChClass(ctx.Buf) == cls {
			if !CursorDec(ctx.Buf) {
				return
			}
			continue
		}
		CursorInc(ctx.Buf)
		break
	}
}

func CmdReplaceChar(ctx Context) func(Context) {
	return func(ctx Context) {
		c := []rune(ctx.Char)
		line := CursorLine(ctx.Buf)
		line.Value[ctx.Buf.Cursor.Char] = c[0]
	}
}

func CmdForwardToChar(_ Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.Mode() == MODE_VISUAL {
			defer SelectionStop(ctx.Buf)
		}

		line := CursorLine(ctx.Buf)
		if line.Value.IsEmpty() {
			return
		}
		for i := ctx.Buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ctx.Char) {
				ctx.Buf.Cursor.Char = i
				ctx.Buf.Cursor.PreserveCharPosition = i
				break
			}
		}
	}
}

func CmdForwardBeforeChar(_ Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.Mode() == MODE_VISUAL {
			defer SelectionStop(ctx.Buf)
		}

		line := CursorLine(ctx.Buf)
		if line.Value.IsEmpty() {
			return
		}
		for i := ctx.Buf.Cursor.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ctx.Char) {
				ctx.Buf.Cursor.Char = i - 1
				ctx.Buf.Cursor.PreserveCharPosition = i - 1
				break
			}
		}
	}
}

func CmdBackwardChar(ctx Context) {
	line := CursorLine(ctx.Buf)
	if len(line.Value) == 0 {
		return
	}

	for i := ctx.Buf.Cursor.Char - 1; i >= 0; i-- {
		if string(line.Value[i]) == ctx.Char {
			ctx.Buf.Cursor.Char = i
			ctx.Buf.Cursor.PreserveCharPosition = i
			break
		}
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

func CmdWindowVSplit(ctx Context) {
	nwin := CreateWindow()
	nwin.VisitBuffer(ctx.Buf)
	ctx.Editor.Windows = append(ctx.Editor.Windows, nwin)
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
	if len(ctx.Editor.Windows) == 1 {
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
	defer func() {
		if ctx.Buf.Selection != nil {
			ctx.Buf.Cursor = ctx.Buf.Selection.Start
		}
		CmdExitInsertMode(ctx)
	}()
	yankSave(ctx)
}

func CmdYankPut(ctx Context) {
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

	yankPut(ctx)
}

func CmdYankPutBefore(ctx Context) {
	if ctx.Editor.Yanks.Len == 0 {
		return
	}

	v := ctx.Editor.Yanks.Last()
	if v.Value.isLine {
		CmdLineOpenAbove(ctx)
		CmdExitInsertMode(ctx)
		yankPut(ctx)
		CmdCursorBeginningOfTheLine(ctx)
	} else {
		yankPut(ctx)
	}
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

func CmdEnsureCursorVisible(ctx Context) {
	_, h := ctx.Editor.View.Size()
	if ctx.Buf.Cursor.Line > ctx.Buf.ScrollOffset+h-minVisibleLines {
		ctx.Buf.ScrollOffset = ctx.Buf.Cursor.Line - h + minVisibleLines
	}

	if ctx.Buf.Cursor.Line < ctx.Buf.ScrollOffset+minVisibleLines {
		ctx.Buf.ScrollOffset = ctx.Buf.Cursor.Line - minVisibleLines
	}
}

func CmdCursorCenter(ctx Context) {
	_, h := ctx.Editor.View.Size()
	ctx.Buf.ScrollOffset = ctx.Buf.Cursor.Line - (h / 2) + minVisibleLines
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

func CmdJumpBack(ctx Context) {
	ctx.Editor.ActiveWindow().Jumps.JumpBack()
	CmdCursorCenter(ctx)
}

func CmdJumpForward(ctx Context) {
	ctx.Editor.ActiveWindow().Jumps.JumpForward()
	CmdCursorCenter(ctx)
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
