package mcwig

import (
	"fmt"
	"slices"
	"strings"
	"text/scanner"
	"unicode"
)

const minVisibleLines = 5

func TextInsert(buf *Buffer, line *Element[Line], pos int, text string) {
	size := len(line.Value)
	if pos >= size {
		pos = size - 1
	}
	if pos < 0 {
		pos = 0
	}

	s := scanner.Scanner{}
	s.Init(strings.NewReader(text))
	s.Whitespace = 0

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		switch tok {
		case '\n':
			prefix := string(line.Value[:pos])
			suffix := string(line.Value[pos:])
			line.Value = []rune(prefix + "\n")
			buf.Lines.insertValueAfter([]rune(suffix), line)
			line = line.Next()
			pos = 0
		default:
			line.Value = slices.Concat(line.Value[:pos], []rune(s.TokenText()), line.Value[pos:])
			pos += len(s.TokenText())
		}
	}
}

func TextDelete(buf *Buffer, selection *Selection) {
	sel := SelectionNormalize(selection)

	lineStart := CursorLineByNum(buf, sel.Start.Line)
	lineEnd := CursorLineByNum(buf, sel.End.Line)

	// if request is to delete more chars then len(end) - we must connect next line
	// since we delete "\n"
	if sel.End.Char >= len(lineEnd.Value) {
		tmpLine := CursorLineByNum(buf, sel.End.Line+1)
		if tmpLine != nil {
			sel.End.Line++
			sel.End.Char = 0
			lineEnd = tmpLine
		}
	}

	if lineStart != lineEnd {
		for lineStart.Next() != lineEnd {
			buf.Lines.Remove(lineStart.Next())
		}
		defer buf.Lines.Remove(lineEnd)
	}

	start := sel.Start.Char
	end := min(len(lineEnd.Value), sel.End.Char)
	lineStart.Value = slices.Concat(lineStart.Value[:start], lineEnd.Value[end:])
}

func CmdEnterInsertMode(ctx Context) {
	line := CursorLine(ctx.Buf)
	if line == nil {
		return
	}

	ctx.Buf.TxStart()

	if len(line.Value) == 0 {
		ctx.Buf.Cursor.Char++
	}

	ctx.Buf.SetMode(MODE_INSERT)
}

func CmdExitInsertMode(ctx Context) {
	defer func() {
		ctx.Buf.SetMode(MODE_NORMAL)
		ctx.Buf.Selection = nil
	}()

	CmdCursorLeft(ctx)
	line := CursorLine(ctx.Buf)
	if ctx.Buf.Cursor.Char >= len(line.Value) {
		CmdGotoLineEnd(ctx)
	}

	ctx.Buf.TxEnd()

	// TODO: this is ugly
	if ctx.Buf.Highlighter != nil {
		ctx.Buf.Highlighter.Build()
	}
}

func CmdInsertModeAfter(ctx Context) {
	CmdCursorRight(ctx)
	CmdEnterInsertMode(ctx)
}

func CmdJoinNextLine(ctx Context) {
	CmdGotoLineEnd(ctx)

	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	line := CursorLine(ctx.Buf)
	TextDelete(ctx.Buf, &Selection{
		Start: Cursor{Line: ctx.Buf.Cursor.Line, Char: len(line.Value) - 1},
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: len(line.Value)},
	})
}

func CmdReplaceChar(ctx Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.TxStart() {
			defer ctx.Buf.TxEnd()
		}
		line := CursorLine(ctx.Buf)
		ctx.Buf.Selection = &Selection{
			Start: ctx.Buf.Cursor,
			End:   ctx.Buf.Cursor,
		}
		SelectionDelete(ctx)
		TextInsert(ctx.Buf, line, ctx.Buf.Cursor.Char, ctx.Char)
	}
}

func CmdDeleteCharForward(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	line := CursorLine(ctx.Buf)
	if len(line.Value) <= 1 {
		return
	}

	if ctx.Buf.Cursor.Char >= len(line.Value)-1 {
		CmdCursorLeft(ctx)
	}

	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   ctx.Buf.Cursor,
	}

	SelectionDelete(ctx)

}

func CmdDeleteCharBackward(ctx Context) {
	if ctx.Buf.Cursor.Char == 0 {
		return
	}
	CmdCursorLeft(ctx)
	CmdDeleteCharForward(ctx)
}

func CmdAppendLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	CmdInsertModeAfter(ctx)
}

func CmdLineOpenBelow(ctx Context) {
	line := CursorLine(ctx.Buf)
	CmdInsertModeAfter(ctx)
	TextInsert(ctx.Buf, line, len(line.Value), "\n")
	CmdCursorLineDown(ctx)
	CmdCursorBeginningOfTheLine(ctx)
	// indent(ctx)
}

func CmdLineOpenAbove(ctx Context) {
	if ctx.Buf.Cursor.Line == 0 {
		CmdEnterInsertMode(ctx)
		TextInsert(ctx.Buf, CursorLine(ctx.Buf), 0, "\n")
		CmdCursorBeginningOfTheLine(ctx)
		return
	}
	CmdCursorLineUp(ctx)
	CmdLineOpenBelow(ctx)
}

func CmdDeleteLine(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	if ctx.Count == 0 {
		ctx.Count = 1
	}

	CmdVisualLineMode(ctx)
	ctx.Buf.Selection.End.Line = min(
		ctx.Buf.Lines.Len-1,
		ctx.Buf.Selection.Start.Line+int(ctx.Count)-1,
	)
	ctx.Buf.Selection.End.Char = len(CursorLineByNum(ctx.Buf, ctx.Buf.Selection.End.Line).Value) - 1
	SelectionDelete(ctx)
	CmdNormalMode(ctx)
}

func CmdDeleteWord(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	_, end := TextObjectWord(ctx.Buf, false)
	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: end},
	}
	SelectionDelete(ctx)
}

func CmdChangeWord(ctx Context) {
	_, end := TextObjectWord(ctx.Buf, false)
	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: end},
	}
	CmdEnterInsertMode(ctx)
	SelectionDelete(ctx)
}

func CmdChangeWORD(ctx Context) {
	start, end := TextObjectWord(ctx.Buf, true)
	ctx.Buf.Cursor.Char = start
	ctx.Buf.Selection = &Selection{
		Start: ctx.Buf.Cursor,
		End:   Cursor{Line: ctx.Buf.Cursor.Line, Char: end},
	}
	CmdEnterInsertMode(ctx)
	SelectionDelete(ctx)
}

func CmdChangeTo(_ Context) func(Context) {
	return func(ctx Context) {
		SelectionStart(ctx.Buf)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf)
		CmdEnterInsertMode(ctx)
		SelectionDelete(ctx)
	}
}

func CmdChangeBefore(_ Context) func(Context) {
	return func(ctx Context) {
		SelectionStart(ctx.Buf)
		CmdForwardBeforeChar(ctx)(ctx)
		SelectionStop(ctx.Buf)
		CmdEnterInsertMode(ctx)
		SelectionDelete(ctx)
	}
}

func CmdChangeEndOfLine(ctx Context) {
	SelectionStart(ctx.Buf)
	CmdGotoLineEnd(ctx)
	SelectionStop(ctx.Buf)
	CmdEnterInsertMode(ctx)
	SelectionDelete(ctx)
}

func CmdChangeLine(ctx Context) {
	line := CursorLine(ctx.Buf)
	idx := 0
	for i, c := range line.Value {
		if !unicode.IsSpace(c) {
			idx = i
			break
		}
	}
	CmdInsertModeAfter(ctx)
	line.Value = line.Value[:idx]
}

func CmdDeleteTo(_ Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.TxStart() {
			defer ctx.Buf.TxEnd()
		}
		SelectionStart(ctx.Buf)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf)
		SelectionDelete(ctx)
	}
}

func CmdDeleteBefore(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	SelectionStart(ctx.Buf)
	CmdForwardBeforeChar(ctx)(ctx)
	SelectionDelete(ctx)
}

func CmdSelectionChange(ctx Context) {
	CmdEnterInsertMode(ctx)
	SelectionDelete(ctx)
}

// TODO: implement correct toggle comment logic: check if all lines are commented - then uncomment.
// else, append comment to each uncommted line.
func CmdToggleComment(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	defer CmdNormalMode(ctx)

	comment := "//"

	cmComment := func(line *Element[Line]) {
		spacePos := 0
		for i, c := range line.Value {
			if !unicode.IsSpace(c) {
				spacePos = i
				break
			}
		}
		tmpData := make([]rune, 0, len(line.Value)+len(comment)+1)
		tmpData = append(tmpData, line.Value[:spacePos]...)
		tmpData = append(tmpData, []rune(comment)...)
		tmpData = append(tmpData, rune(' '))
		tmpData = append(tmpData, line.Value[spacePos:]...)
		line.Value = tmpData

	}

	cmUncomment := func(line *Element[Line], comment string) {
		r := strings.Replace(string(line.Value), comment, "", 1)
		line.Value = []rune(r)
	}

	toggleCommentForLine := func(line *Element[Line]) {
		trimmed := strings.TrimSpace(string(line.Value))
		if strings.HasPrefix(trimmed, comment+" ") {
			cmUncomment(line, comment+" ")
		} else if strings.HasPrefix(trimmed, comment) {
			cmUncomment(line, comment)
		} else {
			cmComment(line)
		}
	}

	if ctx.Buf.Selection != nil {
		selection := SelectionNormalize(ctx.Buf.Selection)

		lineStart := CursorLineByNum(ctx.Buf, selection.Start.Line)
		count := selection.End.Line - selection.Start.Line

		for i := 0; i <= count; i++ {
			line := lineStart
			lineStart = lineStart.Next()
			if len(line.Value) == 0 {
				continue
			}
			toggleCommentForLine(line)
		}
		return
	}

	line := CursorLine(ctx.Buf)
	toggleCommentForLine(line)
}

func CmdSelectionDelete(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	SelectionDelete(ctx)
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

	// cleanup all nodes
	{
		l := ctx.Buf.Lines.First()
		for l != nil {
			next := l.Next()
			l.Value = nil
			ctx.Buf.Lines.Remove(l)
			l = next
		}
		ctx.Buf.Selection = nil
		ctx.Buf.Highlighter = nil
		ctx.Buf.UndoRedo = nil
		ctx.Buf.Tx = nil

	}

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

	ctx.Editor.Buffers = buffers

	if len(buffers) == 0 {
		buf := NewBuffer()
		buf.FilePath = "[No Name]"
		ctx.Editor.Buffers = append(ctx.Editor.Buffers, buf)
		ctx.Editor.ActiveWindow().ShowBuffer(buf)
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
		CmdEnterInsertMode(ctx)
		SelectionDelete(ctx)
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
