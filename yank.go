package mcwig

import "github.com/gdamore/tcell/v2"

type yank struct {
	val    string
	isLine bool
}

type Yanks struct {
	items List[yank]
}

func yankSave(buf *Buffer, line *Element[Line]) {
	var y yank
	if buf.Selection == nil {
		y = yank{string(line.Value), true}
	} else {
		y = yank{SelectionToString(buf), false}
	}
	if buf.Mode == MODE_VISUAL_LINE {
		y.isLine = true
	}

	if buf.Yanks.Len == 0 {
		buf.Yanks.PushBack(y)
		return
	}

	if buf.Yanks.Last().Value != y {
		buf.Yanks.PushBack(y)
	}
}

func yankPut(e *Editor, buf *Buffer) {
	v := buf.Yanks.Last()

	oldMode := buf.Mode
	buf.Mode = MODE_INSERT
	defer func() {
		buf.Mode = oldMode
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
