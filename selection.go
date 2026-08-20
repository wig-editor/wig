package wig

import "strings"

type Selection struct {
	Start Cursor
	End   Cursor
}

func SelectionCursorInRange(sel *Selection, c Cursor) bool {
	s := SelectionNormalize(sel)

	if c.Line < s.Start.Line || c.Line > s.End.Line {
		return false
	}

	if c.Line == s.Start.Line && c.Char < s.Start.Char {
		return false
	}

	if c.Line == s.End.Line && c.Char > s.End.Char {
		return false
	}

	return true
}

func SelectionToString(buf *Buffer, sel *Selection) string {
	if sel == nil {
		return ""
	}

	s := SelectionNormalize(sel)

	lineStart := CursorLineByNum(buf, s.Start.Line)
	lineEnd := CursorLineByNum(buf, s.End.Line)

	if lineStart == nil || lineEnd == nil {
		return ""
	}

	endCh := s.End.Char + 1
	if endCh > len(lineEnd.Value) {
		endCh = len(lineEnd.Value)
	}

	if s.Start.Line == s.End.Line {
		if len(lineStart.Value) == 0 {
			return ""
		}
		return string(lineStart.Value[s.Start.Char:endCh])
	}

	acc := string(lineStart.Value[s.Start.Char:])
	currentLine := lineStart.Next()
	for currentLine != nil {
		if currentLine != lineEnd {
			acc += string(currentLine.Value)
		} else {
			acc += string(currentLine.Value[:endCh])
			break
		}
		currentLine = currentLine.Next()
	}

	return acc
}

// SelectionBlockBounds returns the rectangular bounds of a block selection,
// independent per axis (unlike SelectionNormalize, which couples Line/Char).
func SelectionBlockBounds(sel *Selection) (minLine, maxLine, minChar, maxChar int) {
	minLine, maxLine = sel.Start.Line, sel.End.Line
	if minLine > maxLine {
		minLine, maxLine = maxLine, minLine
	}
	minChar, maxChar = sel.Start.Char, sel.End.Char
	if minChar > maxChar {
		minChar, maxChar = maxChar, minChar
	}
	return
}

// CmdSelectionBlockDelete deletes the rectangular column range [minVisCol,maxVisCol]
// on every line in [minLine,maxLine], independently per line (ragged lines are
// clipped, never merged) — unlike SelectionDelete's single contiguous stream cut.
func CmdSelectionBlockDelete(ctx Context) {
	if ctx.Buf.Selection == nil {
		return
	}
	defer CmdNormalMode(ctx)
	if ctx.Buf.TxStart() {
		defer ctx.Buf.TxEnd()
	}

	sel := ctx.Buf.Selection
	minLine, maxLine := sel.Start.Line, sel.End.Line
	if minLine > maxLine {
		minLine, maxLine = maxLine, minLine
	}

	startLineNode := CursorLineByNum(ctx.Buf, sel.Start.Line)
	endLineNode := CursorLineByNum(ctx.Buf, sel.End.Line)
	startVisCol := VisualCol(startLineNode.Value, sel.Start.Char)
	endVisCol := VisualCol(endLineNode.Value, sel.End.Char)
	minVisCol := min(startVisCol, endVisCol)
	maxVisCol := max(startVisCol, endVisCol)

	for i := minLine; i <= maxLine; i++ {
		line := CursorLineByNum(ctx.Buf, i)
		if line == nil {
			continue
		}
		startChar := RuneIndexFromVisualCol(line.Value, minVisCol)
		endChar := RuneIndexFromVisualCol(line.Value, maxVisCol)

		lineLen := len(line.Value)
		if startChar >= lineLen-1 {
			continue
		}
		if endChar > lineLen-1 {
			endChar = lineLen - 1
		}
		if endChar <= startChar {
			continue
		}

		TextDelete(ctx.Buf, &Selection{
			Start: Cursor{Line: i, Char: startChar},
			End:   Cursor{Line: i, Char: endChar},
		})
	}
	cur := ContextCursorGet(ctx)
	cur.Line = minLine
	cur.Char = RuneIndexFromVisualCol(CursorLineByNum(ctx.Buf, minLine).Value, minVisCol)
	cur.PreserveCharPosition = cur.Char
	ctx.Buf.Selection = nil
}

// CmdSelectionBlockYank captures the rectangular column range [minVisCol,maxVisCol]
// on every line in [minLine,maxLine] as a single block register entry, one
// line per row (independent per-line clipping, no stream join).
func CmdSelectionBlockYank(ctx Context) {
	defer CmdNormalMode(ctx)
	if ctx.Buf.Selection == nil {
		return
	}
	cur := ContextCursorGet(ctx)

	sel := ctx.Buf.Selection
	minLine, maxLine := sel.Start.Line, sel.End.Line
	if minLine > maxLine {
		minLine, maxLine = maxLine, minLine
	}

	startLineNode := CursorLineByNum(ctx.Buf, sel.Start.Line)
	endLineNode := CursorLineByNum(ctx.Buf, sel.End.Line)
	startVisCol := VisualCol(startLineNode.Value, sel.Start.Char)
	endVisCol := VisualCol(endLineNode.Value, sel.End.Char)
	minVisCol := min(startVisCol, endVisCol)
	maxVisCol := max(startVisCol, endVisCol)

	lines := make([]string, 0, maxLine-minLine+1)
	for i := minLine; i <= maxLine; i++ {
		line := CursorLineByNum(ctx.Buf, i)
		if line == nil {
			lines = append(lines, "")
			continue
		}
		startChar := RuneIndexFromVisualCol(line.Value, minVisCol)
		endChar := RuneIndexFromVisualCol(line.Value, maxVisCol)

		lineLen := len(line.Value) - 1 // exclude trailing "\n"
		start := min(startChar, lineLen)
		end := min(endChar, lineLen)
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
	cur.Char = RuneIndexFromVisualCol(CursorLineByNum(ctx.Buf, minLine).Value, minVisCol)
	ctx.Buf.Selection = nil
}

func SelectionNormalize(sel *Selection) Selection {
	if sel == nil {
		return Selection{}
	}

	s := *sel

	if s.Start.Line > s.End.Line {
		s.Start, s.End = s.End, s.Start
	}

	if s.Start.Line == s.End.Line && s.Start.Char > s.End.Char {
		s.Start, s.End = s.End, s.Start
	}

	return s
}

func SelectionStart(buf *Buffer, cur *Cursor) {
	buf.Selection = &Selection{
		Start: *cur,
		End:   *cur,
	}
}

func SelectionStop(buf *Buffer, cur *Cursor) {
	if buf.Selection == nil {
		return
	}
	buf.Selection.End = *cur
}

func WithSelection(fn func(Context)) func(Context) {
	return func(ctx Context) {
		fn(ctx)
		buf := ctx.Buf
		if buf.Selection == nil {
			// TODO: this is workaround for when selection was deleted but did
			// not exited VIS_LINE_MODE
			CmdNormalMode(ctx)
			return
		}
		cur := ContextCursorGet(ctx)
		buf.Selection.End = *cur

		if buf.Mode() == MODE_VISUAL_LINE {
			if buf.Selection.Start.Line > buf.Selection.End.Line {
				lineStart := CursorLineByNum(buf, buf.Selection.Start.Line)
				buf.Selection.Start.Char = len(lineStart.Value) - 1
				buf.Selection.End.Char = 0
			} else {
				lineEnd := CursorLineByNum(buf, buf.Selection.End.Line)
				buf.Selection.Start.Char = 0
				buf.Selection.End.Char = len(lineEnd.Value) - 1
			}
		}
	}
}

func SelectionDelete(ctx Context) {
	if ctx.Buf.Selection == nil {
		return
	}
	defer func() {
		ctx.Buf.Selection = nil
	}()
	sel := SelectionNormalize(ctx.Buf.Selection)
	sel.End.Char++
	TextDelete(ctx.Buf, &sel)
	cur := ContextCursorGet(ctx)
	*cur = sel.Start
}
