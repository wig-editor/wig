package mcwig

import "github.com/gdamore/tcell/v2"

type yank struct {
	val    string
	isLine bool
}

type Yanks struct {
	items List[yank]
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
