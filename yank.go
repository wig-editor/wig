package wig

type yank struct {
	val    string
	isLine bool
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

func yankSave(ctx Context) {
	cur := ContextCursorGet(ctx)
	var y yank
	line := CursorLine(ctx.Buf, cur)
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
	cur := ContextCursorGet(ctx)
	v := ctx.Editor.Yanks.Last()
	TextInsert(ctx.Buf, CursorLine(ctx.Buf, cur), cur.Char, v.Value.val)
	i := len(v.Value.val)
	for i >= 1 {
		i--
		CursorInc(ctx.Buf, cur)
	}
}

