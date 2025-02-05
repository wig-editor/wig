// Movement and window related funcs
package mcwig

import (
	"strings"
	"unicode"
)

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
	if ctx.Buf.Cursor.Char < len(line.Value)-1 {
		ctx.Buf.Cursor.Char++
		ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
	}
}

func CmdCursorLineUp(ctx Context) {
	count := max(ctx.Count, 1)
	ctx.Buf.Cursor.Line = max(
		ctx.Buf.Cursor.Line-int(count),
		0,
	)
	restoreCharPosition(ctx.Buf)
	CmdEnsureCursorVisible(ctx)
}

func CmdCursorLineDown(ctx Context) {
	count := max(ctx.Count, 1)
	ctx.Buf.Cursor.Line = min(
		ctx.Buf.Cursor.Line+int(count),
		ctx.Buf.Lines.Len-1,
	)
	restoreCharPosition(ctx.Buf)
	CmdEnsureCursorVisible(ctx)
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

func CmdGotoLine0(ctx Context) {
	defer CmdEnsureCursorVisible(ctx)
	ctx.Buf.Cursor.Line = min(
		int(ctx.Count),
		ctx.Buf.Lines.Len-1,
	)

	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf)
}

func CmdGotoLineEnd(ctx Context) {
	line := CursorLine(ctx.Buf)
	ctx.Buf.Cursor.Char = len(line.Value) - 2
	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf)
}

func CmdForwardWord(ctx Context) {
	defer CmdEnsureCursorVisible(ctx)

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
	defer CmdEnsureCursorVisible(ctx)

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
