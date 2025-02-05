package mcwig

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

func SelectionToString(buf *Buffer) string {
	if buf.Selection == nil {
		return ""
	}

	s := SelectionNormalize(buf.Selection)

	lineStart := CursorLineByNum(buf, s.Start.Line)
	lineEnd := CursorLineByNum(buf, s.End.Line)

	if lineStart == nil {
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

func SelectionStart(buf *Buffer) {
	buf.Selection = &Selection{
		Start: buf.Cursor,
		End:   buf.Cursor,
	}
}

func SelectionStop(buf *Buffer) {
	buf.Selection.End = buf.Cursor
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
		buf.Selection.End = buf.Cursor

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
	ctx.Buf.Cursor = sel.Start
}
