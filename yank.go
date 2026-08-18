package wig

import (
	"strings"
)

type yank struct {
	val     string
	isLine  bool
	isBlock bool
}
type Yanks struct {
	items List[yank]
}

func CmdYank(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer CmdNormalMode(ctx)
	defer func() {
		if ctx.Buf.Selection != nil {
			cur.Line = ctx.Buf.Selection.Start.Line
			cur.Char = ctx.Buf.Selection.Start.Char
		}
		ctx.Buf.Selection = nil
	}()
	yankSave(ctx)
}

func CmdYankEol(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer CmdNormalMode(ctx)
	defer func() {
		if ctx.Buf.Selection != nil {
			cur.Line = ctx.Buf.Selection.Start.Line
			cur.Char = ctx.Buf.Selection.Start.Char
		}
		ctx.Buf.Selection = nil
	}()
	SelectionStart(ctx.Buf, cur)
	WithSelection(CmdGotoLineEnd)(ctx)
	CmdCursorLeft(ctx)
	SelectionStop(ctx.Buf, cur)
	yankSave(ctx)
}

func CmdYankBeforeChar(ctx Context) func(Context) {
	return func(ctx Context) {
		startCur := *ContextCursorGet(ctx)
		cur := ContextCursorGet(ctx)
		SelectionStart(ctx.Buf, cur)
		CmdForwardBeforeChar(ctx)(ctx)
		SelectionStop(ctx.Buf, cur)
		yankSave(ctx)
		ctx.Buf.Selection = nil
		*cur = startCur
	}
}

// CmdSelectionBlockYank captures the rectangular column range [minChar,maxChar]
// on every line in [minLine,maxLine] as a single block register entry, one
// line per row (independent per-line clipping, no stream join).
func CmdSelectionBlockYank(ctx Context) {
	defer CmdNormalMode(ctx)
	if ctx.Buf.Selection == nil {
		return
	}
	cur := ContextCursorGet(ctx)
	minLine, maxLine, minChar, maxChar := SelectionBlockBounds(ctx.Buf.Selection)
	lines := make([]string, 0, maxLine-minLine+1)
	for i := minLine; i <= maxLine; i++ {
		line := CursorLineByNum(ctx.Buf, i)
		if line == nil {
			lines = append(lines, "")
			continue
		}
		lineLen := len(line.Value) - 1 // exclude trailing "\n"
		start := min(minChar, lineLen)
		end := min(maxChar+1, lineLen)
		if end < start {
			end = start
		}
		lines = append(lines, string(line.Value[start:end]))
	}
	y := yank{val: strings.Join(lines, "\n"), isBlock: true}
	if ctx.Editor.Yanks.Len == 0 || ctx.Editor.Yanks.Last().Value != y {
		ctx.Editor.Yanks.PushBack(y)
	}
	cur.Line = minLine
	cur.Char = minChar
	ctx.Buf.Selection = nil
}
func CmdYankToChar(_ Context) func(Context) {
	return func(ctx Context) {
		startCur := *ContextCursorGet(ctx)
		cur := ContextCursorGet(ctx)
		SelectionStart(ctx.Buf, cur)
		CmdForwardToChar(ctx)(ctx)
		SelectionStop(ctx.Buf, cur)
		yankSave(ctx)
		ctx.Buf.Selection = nil
		*cur = startCur
	}
}

func CmdYankPut(ctx Context) {
	cur := ContextCursorGet(ctx)
	if ctx.Editor.Yanks.Len == 0 {
		return
	}
	if ctx.Editor.Yanks.Last().Value.isBlock {
		blockPut(ctx, false)
		return
	}
	if ctx.Buf.Selection != nil {
		if ctx.Buf.TxStart() {
			defer yankSave(ctx, SelectionToString(ctx.Buf, ctx.Buf.Selection))
			if ctx.Buf.Mode() == MODE_VISUAL {
				SelectionDelete(ctx)
			}
			if ctx.Buf.Mode() == MODE_VISUAL_LINE {
				SelectionDelete(ctx)
				CmdCursorLineUp(ctx)
				line := CursorLine(ctx.Buf, cur)
				CmdAppendLine(ctx)
				TextInsert(ctx.Buf, line, len(line.Value)-1, "\n")
			}
			ctx.Buf.TxEnd()
		}
	}

	v := ctx.Editor.Yanks.Last()
	if v.Value.isLine {
		CmdCursorLineDown(ctx)
		CmdYankPutBefore(ctx)
		return
	}

	CmdEnterInsertMode(ctx)
	defer CmdExitInsertMode(ctx)

	CmdCursorRight(ctx)
	yankPut(ctx)
}

func CmdYankPutBefore(ctx Context) {
	if ctx.Editor.Yanks.Len == 0 {
		return
	}
	if ctx.Editor.Yanks.Last().Value.isBlock {
		blockPut(ctx, true)
		return
	}
	cur := ContextCursorGet(ctx)
	CmdEnterInsertMode(ctx)
	defer CmdExitInsertMode(ctx)

	v := ctx.Editor.Yanks.Last()
	if v.Value.isLine {
		CmdLineOpenAbove(ctx)
		CmdCursorBeginningOfTheLine(ctx)

		// clear any indentation
		SelectionStart(ctx.Buf, cur)
		CmdGotoLineEnd(ctx)
		SelectionStop(ctx.Buf, cur)
		SelectionDelete(ctx)

		yankPut(ctx)
	} else {
		yankPut(ctx)
	}
}

func yankSave(ctx Context, text ...string) {
	cur := ContextCursorGet(ctx)
	var y yank
	line := CursorLine(ctx.Buf, cur)

	if len(text) == 0 {
		if ctx.Buf.Selection == nil {
			y = yank{val: string(line.Value)}
		} else {
			st := SelectionToString(ctx.Buf, ctx.Buf.Selection)
			if len(st) == 0 {
				return
			}
			y = yank{val: st}
		}
	} else {
		y = yank{val: text[0]}
	}

	y.isLine = (ctx.Buf.Mode() == MODE_VISUAL_LINE) || ctx.Buf.Selection == nil

	if ctx.Editor.Yanks.Len == 0 {
		ctx.Editor.Yanks.PushBack(y)
		return
	}

	if ctx.Editor.Yanks.Last().Value != y {
		ctx.Editor.Yanks.PushBack(y)
	}
}

func yankPut(ctx Context) {
	cur := ContextCursorGet(ctx)
	v := ctx.Editor.Yanks.Last()
	TextInsert(ctx.Buf, CursorLine(ctx.Buf, cur), cur.Char, v.Value.val)
	i := len(v.Value.val)
	for i >= 1 {
		i--
		CursorInc(ctx.Buf, cur)
	}
}

// blockPut inserts a blockwise register at a fixed column across successive
// lines (one register line per buffer line), padding short lines with spaces
// and appending new lines at EOF if the block extends past the last line —
// unlike yankPut's single flat stream insert.
func blockPut(ctx Context, before bool) {
	cur := ContextCursorGet(ctx)
	v := ctx.Editor.Yanks.Last()
	lines := strings.Split(v.Value.val, "\n")
	startChar := cur.Char
	if !before {
		startChar++
	}
	startLine := cur.Line
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}
	for i, text := range lines {
		lineNum := startLine + i
		line := CursorLineByNum(ctx.Buf, lineNum)
		if line == nil {
			last := ctx.Buf.Lines.Last()
			TextInsert(ctx.Buf, last, len(last.Value)-1, "\n"+strings.Repeat(" ", startChar)+text)
			continue
		}
		lineLen := len(line.Value) - 1 // exclude trailing "\n"
		if startChar > lineLen {
			TextInsert(ctx.Buf, line, lineLen, strings.Repeat(" ", startChar-lineLen))
			line = CursorLineByNum(ctx.Buf, lineNum)
		}
		TextInsert(ctx.Buf, line, startChar, text)
	}
	cur.Line = startLine
	cur.Char = startChar
}
