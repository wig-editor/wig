package wig

import (
	"fmt"
	"slices"
	"strings"
	"text/scanner"
	"unicode"
)

const minVisibleLines = 5
const smode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments

func TextInsert(buf *Buffer, line *Element[Line], pos int, text string) {
	sline := CursorNumByLine(buf, line)

	event := EventTextChange{
		Buf:    buf,
		Start:  Position{Line: sline, Char: pos},
		End:    Position{Line: sline, Char: pos},
		NewEnd: Position{Line: sline, Char: pos},
		Text:   text,
	}
	if pos < 0 {
		pos = 0
	}
	size := len(line.Value)
	if pos >= size {
		pos = size - 1
	}

	s := scanner.Scanner{}
	s.Init(strings.NewReader(text))
	s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<' '
	s.Mode = smode

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		switch tok {
		case '\n':
			prefix := string(line.Value[:pos])
			suffix := string(line.Value[pos:])
			line.Value = []rune(prefix + "\n")
			buf.Lines.insertValueAfter([]rune(suffix), line)
			line = line.Next()
			pos = 0
			event.NewEnd.Line++
			event.NewEnd.Char = 0
		default:
			line.Value = slices.Concat(line.Value[:pos], []rune(s.TokenText()), line.Value[pos:])
			pos += len(s.TokenText())
			event.NewEnd.Char++
		}
	}

	EditorInst.Events.Broadcast(event)
}

func TextDelete(buf *Buffer, selection *Selection) {
	defer func() {
		if buf.Lines.Len == 1 && len(buf.Lines.First().Value) == 0 {
			buf.Lines.First().Value = []rune{'\n'}
		}
	}()

	sel := SelectionNormalize(selection)
	lineStart := CursorLineByNum(buf, sel.Start.Line)
	lineEnd := CursorLineByNum(buf, sel.End.Line)
	sel.End.Char--
	oldText := SelectionToString(buf, &sel)
	sel.End.Char++

	// if request is to delete more chars then len(end) - we must connect next line
	// since we delete "\n"
	if sel.End.Char >= len(lineEnd.Value) {
		tmpLine := lineEnd.Next()
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
		buf.Lines.Remove(lineEnd)
	}

	start := max(0, sel.Start.Char)
	end := min(len(lineEnd.Value), sel.End.Char)

	lineStart.Value = slices.Concat(lineStart.Value[:start], lineEnd.Value[end:])

	event := EventTextChange{
		Buf:     buf,
		Start:   Position{Line: sel.Start.Line, Char: sel.Start.Char},
		End:     Position{Line: sel.End.Line, Char: sel.End.Char},
		Text:    "",
		OldText: oldText,
	}

	EditorInst.Events.Broadcast(event)
}

func CmdJoinNextLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	cur := ContextCursorGet(ctx)

	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	line := CursorLine(ctx.Buf, cur)
	TextDelete(ctx.Buf, &Selection{
		Start: Cursor{Line: cur.Line, Char: len(line.Value) - 1},
		End:   Cursor{Line: cur.Line, Char: len(line.Value)},
	})
}

func CmdReplaceChar(ctx Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.TxStart() {
			defer ctx.Buf.TxEnd()
		}
		cur := ContextCursorGet(ctx)
		line := CursorLine(ctx.Buf, cur)
		ctx.Buf.Selection = &Selection{
			Start: *cur,
			End:   *cur,
		}
		SelectionDelete(ctx)
		TextInsert(ctx.Buf, line, cur.Char, ctx.Char)
	}
}

func CmdDeleteCharForward(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	if len(line.Value) <= 1 {
		return
	}

	if cur.Char >= len(line.Value)-1 {
		CmdCursorLeft(ctx)
	}

	ctx.Buf.Selection = &Selection{
		Start: *cur,
		End:   *cur,
	}

	SelectionDelete(ctx)
}

func CmdDeleteCharBackward(ctx Context) {
	cur := ContextCursorGet(ctx)
	if cur.Char == 0 {
		return
	}
	CmdCursorLeft(ctx)
	CmdDeleteCharForward(ctx)
}

func CmdAppendLine(ctx Context) {
	CmdGotoLineEnd(ctx)
	CmdEnterInsertModeAppend(ctx)
}

func CmdLineOpenBelow(ctx Context) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	CmdAppendLine(ctx)
	TextInsert(ctx.Buf, line, len(line.Value)-1, "\n")
	CmdCursorLineDown(ctx)
	CmdCursorBeginningOfTheLine(ctx)
	indent(ctx)
}

func CmdLineOpenAbove(ctx Context) {
	cur := ContextCursorGet(ctx)
	if cur.Line == 0 {
		CmdEnterInsertMode(ctx)
		TextInsert(ctx.Buf, CursorLine(ctx.Buf, cur), 0, "\n")
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
	yankSave(ctx)
	SelectionDelete(ctx)
	CmdNormalMode(ctx)
}

func CmdDeleteWord(ctx Context) {
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	cur := ContextCursorGet(ctx)
	_, end := TextObjectWord(ctx, false)
	ctx.Buf.Selection = &Selection{
		Start: *cur,
		End:   Cursor{Line: cur.Line, Char: end},
	}
	yankSave(ctx)
	SelectionDelete(ctx)
}

func CmdChangeWord(ctx Context) {
	cur := ContextCursorGet(ctx)
	_, end := TextObjectWord(ctx, false)
	ctx.Buf.Selection = &Selection{
		Start: *cur,
		End:   Cursor{Line: cur.Line, Char: end},
	}
	CmdEnterInsertMode(ctx)
	yankSave(ctx)
	SelectionDelete(ctx)
}

func CmdChangeWORD(ctx Context) {
	cur := ContextCursorGet(ctx)
	start, end := TextObjectWord(ctx, true)
	cur.Char = start
	ctx.Buf.Selection = &Selection{
		Start: *cur,
		End:   Cursor{Line: cur.Line, Char: end},
	}
	CmdEnterInsertMode(ctx)
	yankSave(ctx)
	SelectionDelete(ctx)
}

func CmdChangeTo(_ Context) func(Context) {
	return func(ctx Context) {
		cur := ContextCursorGet(ctx)
		SelectionStart(ctx.Buf, cur)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf, cur)
		CmdEnterInsertMode(ctx)
		yankSave(ctx)
		SelectionDelete(ctx)
	}
}

func CmdChangeBefore(_ Context) func(Context) {
	return func(ctx Context) {
		cur := ContextCursorGet(ctx)
		SelectionStart(ctx.Buf, cur)
		CmdForwardBeforeChar(ctx)(ctx)
		SelectionStop(ctx.Buf, cur)
		CmdEnterInsertMode(ctx)
		yankSave(ctx)
		SelectionDelete(ctx)
	}
}

func CmdChangeLine(ctx Context) {
	CmdCursorFirstNonBlank(ctx)
	CmdChangeEndOfLine(ctx)
}

func CmdChangeEndOfLine(ctx Context) {
	ctx.Char = "\n"
	CmdChangeBefore(ctx)(ctx)
}

func CmdDeleteEndOfLine(ctx Context) {
	ctx.Char = "\n"
	CmdChangeBefore(ctx)(ctx)
	CmdNormalMode(ctx)
}

func CmdDeleteTo(_ Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.TxStart() {
			defer ctx.Buf.TxEnd()
		}
		cur := ContextCursorGet(ctx)
		SelectionStart(ctx.Buf, cur)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf, cur)
		yankSave(ctx)
		SelectionDelete(ctx)
	}
}

func CmdDeleteBefore(ctx Context) func(Context) {
	return func(ctx Context) {
		if ctx.Buf.TxStart() {
			defer ctx.Buf.TxEnd()
		}
		cur := ContextCursorGet(ctx)
		SelectionStart(ctx.Buf, cur)
		CmdForwardBeforeChar(ctx)(ctx)
		SelectionStop(ctx.Buf, cur)
		yankSave(ctx)
		SelectionDelete(ctx)
	}
}

func CmdSelectionChange(ctx Context) {
	CmdEnterInsertMode(ctx)
	yankSave(ctx)
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

	// TODO: improve. make comments like all other normal editors!
	cmComment := func(line *Element[Line]) {
		spacePos := 0
		for i, c := range line.Value {
			if !unicode.IsSpace(c) {
				spacePos = i
				break
			}
		}
		TextInsert(ctx.Buf, line, spacePos, comment+" ")
	}

	cmUncomment := func(line *Element[Line], comment string) {
		r := strings.Replace(string(line.Value), comment, "", 1)
		lineNum := CursorNumByLine(ctx.Buf, line)
		TextDelete(ctx.Buf, &Selection{
			Start: Cursor{Line: lineNum, Char: 0},
			End:   Cursor{Line: lineNum, Char: len(line.Value) - 1},
		})
		TextInsert(ctx.Buf, line, 0, r[:len(r)-1])
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
			if line.Value.IsEmpty() {
				continue
			}
			toggleCommentForLine(line)
		}
		return
	}

	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	toggleCommentForLine(line)
}

func CmdSelectionDelete(ctx Context) {
	defer CmdNormalMode(ctx)
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	yankSave(ctx)
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

	for i, b := range buffers {
		if b == ctx.Buf {
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

			buffers = slices.Delete(buffers, i, i+1)

			if len(buffers) > 0 {
				idx := i - 1
				idx = max(idx, 0)
				ctx.Buf = buffers[idx]
				cur := ContextCursorGet(ctx)
				ctx.Editor.ActiveWindow().VisitBuffer(ctx, *cur)
			}

			ctx.Editor.Windows = slices.DeleteFunc(ctx.Editor.Windows, func(win *Window) bool {
				if win.buf == b {
					return true
				}
				return false
			})
		}
	}

	ctx.Editor.Buffers = buffers

	if len(buffers) == 0 {
		CmdNewBuffer(ctx)
	}

	ctx.Editor.Lsp.DidClose(ctx.Buf)
}

func CmdNewBuffer(ctx Context) {
	buf := NewBuffer()
	buf.Lines.PushBack([]rune("\n"))
	buf.FilePath = "[No Name]"
	ctx.Editor.Buffers = append(ctx.Editor.Buffers, buf)
	ctx.Editor.ActiveWindow().ShowBuffer(buf)
}

func CmdIndentOrComplete(ctx Context) {
	ctx.Editor.Lsp.Completion(ctx.Buf)
}

func CmdChangeInsideBlock(ctx Context) {
	switch ctx.Char {
	case "w":
		CmdChangeWORD(ctx)
	case "(", "[", "{", "'", "\"":
		ctx.Editor.EchoMessage("TODO: rewrite this")
		found, sel, _ := TextObjectBlock(ctx.Buf, rune(ctx.Char[0]), false) // TODO: handle unicode
		if !found {
			return
		}
		ctx.Buf.Selection = sel
		// ctx.Buf.Cursor = cur
		CmdEnterInsertMode(ctx)
		SelectionDelete(ctx)
	}
}

func CmdUndo(ctx Context) {
	ctx.Buf.UndoRedo.Undo()
	// TODO: undo/redo does not support lsp change text events.
	// so we simply reload how buffer for now.
	{
		ctx.Editor.Lsp.DidClose(ctx.Buf)
		ctx.Editor.Lsp.DidOpen(ctx.Buf)
	}
}

func CmdRedo(ctx Context) {
	ctx.Buf.UndoRedo.Redo()
	{
		ctx.Editor.Lsp.DidClose(ctx.Buf)
		ctx.Editor.Lsp.DidOpen(ctx.Buf)
	}
}

func CmdExit(ctx Context) {
	ctx.Editor.ExitCh <- 1
}

func CmdEnterInsertMode(ctx Context) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	if line == nil {
		return
	}
	ctx.Buf.TxStart()
	setBufferMode(ctx, MODE_INSERT)
}

func CmdEnterInsertModeAppend(ctx Context) {
	CmdCursorRight(ctx)
	CmdEnterInsertMode(ctx)
}

func CmdVisualMode(ctx Context) {
	cur := ContextCursorGet(ctx)
	SelectionStart(ctx.Buf, cur)
	setBufferMode(ctx, MODE_VISUAL)
}

func CmdExitInsertMode(ctx Context) {
	CmdNormalMode(ctx)
}

func CmdNormalMode(ctx Context) {
	if ctx.Buf.Mode() == MODE_INSERT {
		cur := ContextCursorGet(ctx)
		line := CursorLine(ctx.Buf, cur)
		CmdCursorLeft(ctx)
		if cur.Char >= len(line.Value) {
			CmdGotoLineEnd(ctx)
		}
	}

	ctx.Buf.TxEnd()
	setBufferMode(ctx, MODE_NORMAL)
	ctx.Buf.Selection = nil
}

func CmdVisualLineMode(ctx Context) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	SelectionStart(ctx.Buf, cur)
	ctx.Buf.Selection.Start.Char = 0
	ctx.Buf.Selection.End.Char = len(line.Value) - 1
	setBufferMode(ctx, MODE_VISUAL_LINE)
}

func setBufferMode(ctx Context, newMode Mode) {
	// ctx.Editor.Events.Broadcast(EventBufferModeChange{
	// Buf:     ctx.Buf,
	// OldMode: ctx.Buf.Mode(),
	// NewMode: newMode,
	// })
	ctx.Buf.SetMode(newMode)
}

func CmdMacroRecord(ctx Context) func(Context) {
	if ctx.Editor.Keys.Macros.Recording() {
		ctx.Editor.Keys.Macros.Stop()
		ctx.Editor.Keys.resetState()
		return nil
	}

	return func(ctx Context) {
		ctx.Editor.Keys.Macros.Start(ctx.Char)
	}
}

func CmdMacroPlay(ctx Context) func(Context) {
	return func(ctx Context) {
		ctx.Editor.Keys.resetState()

		reg := ctx.Char
		count := max(ctx.Count, 1)
		for i := uint32(0); i < count; i++ {
			ctx.Editor.Keys.Macros.Play(reg)
		}
	}
}

func CmdMacroRepeat(ctx Context) {
	ctx.Editor.Keys.Macros.Play(".")
}

func CmdAutocompleteTrigger(ctx Context) {
	ctx.Editor.AutocompleteTrigger(ctx)
}

