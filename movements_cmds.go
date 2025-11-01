// Movement and window related funcs
package wig

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

func CmdScrollUp(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer func() {
		if cur.ScrollOffset < 0 {
			cur.ScrollOffset = 0
		}
	}()
	if cur.ScrollOffset > 0 {
		cur.ScrollOffset--

		_, h := ctx.Editor.View.Size()
		if cur.Line > cur.ScrollOffset+h-minVisibleLines {
			CmdCursorLineUp(ctx)
		}
	}
}

func CmdScrollDown(ctx Context) {
	cur := ContextCursorGet(ctx)
	if cur.ScrollOffset < ctx.Buf.Lines.Len-minVisibleLines {
		cur.ScrollOffset++

		if cur.Line <= cur.ScrollOffset+minVisibleLines {
			CmdCursorLineDown(ctx)
		}
	}
}

func CmdCursorLeft(ctx Context) {
	count := max(ctx.Count, 1)
	cur := ContextCursorGet(ctx)

	for i := uint32(0); i < count; i++ {
		if cur.Char > 0 {
			cur.Char--
			cur.PreserveCharPosition = cur.Char
		}
	}
}

func CmdCursorRight(ctx Context) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)

	count := max(ctx.Count, 1)

	for i := uint32(0); i < count; i++ {
		if cur.Char < len(line.Value)-1 {
			cur.Char++
			cur.PreserveCharPosition = cur.Char
		}
	}
}

func CmdCursorLineUp(ctx Context) {
	count := max(ctx.Count, 1)
	cur := ContextCursorGet(ctx)
	cur.Line = max(
		cur.Line-int(count),
		0,
	)
	restoreCharPosition(ctx.Buf, cur)
	CmdEnsureCursorVisible(ctx)
}

func CmdCursorLineDown(ctx Context) {
	cur := ContextCursorGet(ctx)
	count := max(ctx.Count, 1)
	cur.Line = min(
		cur.Line+int(count),
		ctx.Buf.Lines.Len-1,
	)

	restoreCharPosition(ctx.Buf, cur)
	CmdEnsureCursorVisible(ctx)
}

func CmdCursorBeginningOfTheLine(ctx Context) {
	cur := ContextCursorGet(ctx)
	cur.Char = 0
	cur.PreserveCharPosition = 0
}

func CmdCursorFirstNonBlank(ctx Context) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	if line.Value.IsEmpty() {
		CmdGotoLineEnd(ctx)
		return
	}
	CmdCursorBeginningOfTheLine(ctx)
	if len(line.Value) <= 1 {
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
	count := max(ctx.Count, 1)
	cur := ContextCursorGet(ctx)
	defer CmdEnsureCursorVisible(ctx)
	cur.Line = min(int(count)-1, ctx.Buf.Lines.Len-1)
	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf, cur)
}

func CmdGotoLineEndOfFile(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer CmdEnsureCursorVisible(ctx)
	cur.Line = ctx.Buf.Lines.Len - 1
	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf, cur)
}

func CmdGotoLineEnd(ctx Context) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	cur.Char = len(line.Value) - 1
	ctx.Editor.ActiveWindow().Jumps.Push(ctx.Buf, cur)
}

func CmdForwardWord(ctx Context) {
	defer CmdEnsureCursorVisible(ctx)
	cur := ContextCursorGet(ctx)

	// on line change skip all whitespaces
	startLine := cur.Char
	defer func() {
		if startLine != cur.Line {
			for CursorChClass(ctx.Buf, cur) == chWhitespace {
				if !CursorInc(ctx.Buf, cur) {
					return
				}
			}
		}
	}()

	count := max(ctx.Count, 1)

	for i := uint32(0); i < count; i++ {
		line := CursorLine(ctx.Buf, cur)
		cls := CursorChClass(ctx.Buf, cur)
		CursorInc(ctx.Buf, cur)

		// return on line change
		if line != CursorLine(ctx.Buf, cur) {
			return
		}

		if cls != chWhitespace {
			for CursorChClass(ctx.Buf, cur) == cls {
				if !CursorInc(ctx.Buf, cur) {
					return
				}
			}
		}

		// skip whitespace
		line = CursorLine(ctx.Buf, cur)
		for CursorChClass(ctx.Buf, cur) == chWhitespace {
			if !CursorInc(ctx.Buf, cur) {
				return
			}
			if line != CursorLine(ctx.Buf, cur) {
				return
			}
		}
	}
}

func CmdBackwardWord(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer CmdEnsureCursorVisible(ctx)
	defer CursorInc(ctx.Buf, cur)

	cls := CursorChClass(ctx.Buf, cur)
	CursorDec(ctx.Buf, cur)

	count := max(ctx.Count, 1)

	for i := uint32(0); i < count; i++ {
		for CursorChClass(ctx.Buf, cur) == chWhitespace {
			if !CursorDec(ctx.Buf, cur) {
				return
			}
		}

		cls = CursorChClass(ctx.Buf, cur)
		for {
			if CursorChClass(ctx.Buf, cur) == cls {
				if !CursorDec(ctx.Buf, cur) {
					return
				}
				continue
			}
			break
		}
	}
}

func CmdForwardBeforeChar(_ Context) func(Context) {
	return func(ctx Context) {
		cur := ContextCursorGet(ctx)
		if ctx.Buf.Mode() == MODE_VISUAL {
			defer SelectionStop(ctx.Buf, cur)
		}

		line := CursorLine(ctx.Buf, cur)
		if line.Value.IsEmpty() {
			return
		}
		for i := cur.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ctx.Char) {
				cur.Char = i - 1
				cur.PreserveCharPosition = i - 1
				break
			}
		}
	}
}

func CmdForwardToChar(_ Context) func(Context) {
	return func(ctx Context) {
		cur := ContextCursorGet(ctx)
		if ctx.Buf.Mode() == MODE_VISUAL {
			defer SelectionStop(ctx.Buf, cur)
		}

		line := CursorLine(ctx.Buf, cur)
		if line.Value.IsEmpty() {
			return
		}
		for i := cur.Char + 1; i < len(line.Value); i++ {
			if strings.EqualFold(string(line.Value[i]), ctx.Char) {
				cur.Char = i
				cur.PreserveCharPosition = i
				break
			}
		}
	}
}

func CmdBackwardChar(ctx Context) func(Context) {
	return func(ctx Context) {
		cur := ContextCursorGet(ctx)
		line := CursorLine(ctx.Buf, cur)
		if len(line.Value) == 0 {
			return
		}

		for i := cur.Char - 1; i >= 0; i-- {
			if string(line.Value[i]) == ctx.Char {
				cur.Char = i
				cur.PreserveCharPosition = i
				break
			}
		}
	}
}

func CmdWindowVSplit(ctx Context) {
	cur := ContextCursorGet(ctx)
	nwin := CreateWindow(ctx.Editor.ActiveWindow())
	nwin.VisitBuffer(ctx, *cur)
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

func CmdWindowCloseOther(ctx Context) {
	if len(ctx.Editor.Windows) == 1 {
		return
	}

	curWin := ctx.Editor.ActiveWindow()
	ctx.Editor.Windows = slices.DeleteFunc(ctx.Editor.Windows, func(win *Window) bool {
		if curWin == win {
			return false
		}
		return true
	})
}

func CmdWindowCloseAndKillBuffer(ctx Context) {
	CmdKillBuffer(ctx)
	CmdWindowClose(ctx)
}

func CmdEnsureCursorVisible(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer func() {
		if cur.ScrollOffset < 0 {
			cur.ScrollOffset = 0
		}
	}()

	_, h := ctx.Editor.View.Size()
	if cur.Line > cur.ScrollOffset+h-minVisibleLines {
		cur.ScrollOffset = cur.Line - h + minVisibleLines
	}

	if cur.Line < cur.ScrollOffset+minVisibleLines {
		cur.ScrollOffset = cur.Line - minVisibleLines
	}
}

func CmdCursorCenter(ctx Context) {
	cur := ContextCursorGet(ctx)
	defer func() {
		if cur.ScrollOffset < 0 {
			cur.ScrollOffset = 0
		}
	}()

	_, h := ctx.Editor.View.Size()
	cur.ScrollOffset = cur.Line - (h / 2) + minVisibleLines
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
	if last == nil {
		return
	}

	getPrev := func() string {
		for item := last; item != nil; item = item.Prev() {
			if item.Value.FilePath != last.Value.FilePath {
				return item.Value.FilePath
			}
		}
		return ""
	}

	prev := getPrev()
	fmt.Println(prev)
	if prev == "" {
		return
	}

	var b *Buffer
	if last.Value.FilePath == ctx.Editor.ActiveWindow().Buffer().FilePath {
		b = ctx.Editor.BufferFindByFilePath(prev, false)
	} else {
		b = ctx.Editor.BufferFindByFilePath(last.Value.FilePath, false)
	}

	ctx.Editor.ActiveWindow().ShowBuffer(b)
}

