package mcwig

import "github.com/gdamore/tcell/v2"

type yank struct {
	val    string
	isLine bool
}

type Yanks struct {
	items List[yank]
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
		newLine(ctx.Buf, CursorLine(ctx.Buf))
		CmdCursorLineDown(ctx)
		CmdCursorBeginningOfTheLine(ctx)
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

func yankSave(ctx Context) {
	var y yank
	line := CursorLine(ctx.Buf)
	if ctx.Buf.Selection == nil {
		y = yank{string(line.Value), true}
	} else {
		st := SelectionToString(ctx.Buf)
		if len(st) == 0 {
			return
		}
		y = yank{st, false}
	}
	if ctx.Buf.Mode() == MODE_VISUAL_LINE {
		y.isLine = true
	}

	if ctx.Editor.Yanks.Len == 0 {
		ctx.Editor.Yanks.PushBack(y)
		return
	}

	if ctx.Editor.Yanks.Last().Value != y {
		ctx.Editor.Yanks.PushBack(y)
	}
}

func yankPut(ctx Context) {
	v := ctx.Editor.Yanks.Last()

	oldMode := ctx.Buf.Mode()
	ctx.Buf.SetMode(MODE_INSERT)
	defer func() {
		ctx.Buf.SetMode(oldMode)
	}()

	// TODO: use Selection:Replace/Change
	r := []rune(v.Value.val)
	for _, ch := range r {
		k := tcell.KeyRune
		if ch == '\n' {
			k = tcell.KeyEnter
		}
		HandleInsertKey(ctx, tcell.NewEventKey(k, ch, tcell.ModNone))
	}
}
