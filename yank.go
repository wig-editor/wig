package wig

type yank struct {
	val    string
	isLine bool
}

type Yanks struct {
	items List[yank]
}

func CmdYank(ctx Context) {
	defer CmdNormalMode(ctx)
	defer func() {
		if ctx.Buf.Selection != nil {
			ctx.Buf.Cursor = ctx.Buf.Selection.Start
		}
		ctx.Buf.Selection = nil
	}()
	yankSave(ctx)
}

func CmdYankEol(ctx Context) {
	defer CmdNormalMode(ctx)
	defer func() {
		if ctx.Buf.Selection != nil {
			ctx.Buf.Cursor = ctx.Buf.Selection.Start
		}
		ctx.Buf.Selection = nil
	}()
	SelectionStart(ctx.Buf)
	WithSelection(CmdGotoLineEnd)(ctx)
	CmdCursorLeft(ctx)
	SelectionStop(ctx.Buf)
	yankSave(ctx)
}

func CmdYankPut(ctx Context) {
	if ctx.Editor.Yanks.Len == 0 {
		return
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

	CmdEnterInsertMode(ctx)
	defer CmdExitInsertMode(ctx)

	v := ctx.Editor.Yanks.Last()
	if v.Value.isLine {
		CmdLineOpenAbove(ctx)
		CmdCursorBeginningOfTheLine(ctx)

		// clear any indentation
		SelectionStart(ctx.Buf)
		CmdGotoLineEnd(ctx)
		SelectionStop(ctx.Buf)
		SelectionDelete(ctx)

		yankPut(ctx)
	} else {
		yankPut(ctx)
	}
}

func yankSave(ctx Context) {
	var y yank
	line := CursorLine(ctx.Buf)
	if ctx.Buf.Selection == nil {
		y = yank{val: string(line.Value)}
	} else {
		st := SelectionToString(ctx.Buf, ctx.Buf.Selection)
		if len(st) == 0 {
			return
		}
		y = yank{val: st}
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
	v := ctx.Editor.Yanks.Last()
	TextInsert(ctx.Buf, CursorLine(ctx.Buf), ctx.Buf.Cursor.Char, v.Value.val)
	i := len(v.Value.val)
	for i >= 1 {
		i--
		CursorInc(ctx.Buf)
	}
}

