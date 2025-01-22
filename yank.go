package mcwig

import "github.com/gdamore/tcell/v2"

type yank struct {
	val    string
	isLine bool
}

type Yanks struct {
	items List[yank]
}

func yankSave(e *Editor, buf *Buffer, line *Element[Line]) {
	var y yank
	if buf.Selection == nil {
		y = yank{string(line.Value), true}
	} else {
		st := SelectionToString(buf)
		if len(st) == 0 {
			return
		}
		y = yank{st, false}
	}
	if buf.Mode() == MODE_VISUAL_LINE {
		y.isLine = true
	}

	if e.Yanks.Len == 0 {
		e.Yanks.PushBack(y)
		return
	}

	if e.Yanks.Last().Value != y {
		e.Yanks.PushBack(y)
	}
}

func yankPut(e *Editor, buf *Buffer) {
	v := e.Yanks.Last()

	oldMode := buf.Mode()
	buf.SetMode(MODE_INSERT)
	defer func() {
		buf.SetMode(oldMode)
	}()

	r := []rune(v.Value.val)
	for _, ch := range r {
		k := tcell.KeyRune
		if ch == '\n' {
			k = tcell.KeyEnter
		}
		HandleInsertKey(e, tcell.NewEventKey(k, ch, tcell.ModNone))
	}
}
