package wig

type Window struct {
	buf     *Buffer // active buffer
	cursors map[*Buffer]*Cursor

	Jumps *Jumps
}

// Jump to buffer and location. Records jump history.
func (win *Window) VisitBuffer(ctx Context, cursor ...Cursor) {
	if ctx.Buf == nil {
		return
	}

	cur := WindowCursorGet(win, win.buf)
	if win.buf != nil {
		win.Jumps.Push(win.buf, cur)
	}

	if len(cursor) > 0 {
		newCur := &Cursor{}
		newCur.Line = cursor[0].Line
		newCur.Char = cursor[0].Char
		win.cursors[ctx.Buf] = newCur
	}

	win.Jumps.Push(ctx.Buf, cur)
	win.buf = ctx.Buf

	CmdCursorCenter(ctx)
}

// Show buffer. No history.
func (win *Window) ShowBuffer(buf *Buffer) {
	if buf != nil {
		win.buf = buf
	}
}

func (win *Window) Buffer() *Buffer {
	return win.buf
}

// Specify parent window to inherit cursors
func CreateWindow(parent *Window) *Window {
	cursors := map[*Buffer]*Cursor{}
	if parent != nil {
		for k, v := range parent.cursors {
			nc := *v
			cursors[k] = &nc
		}
	}

	return &Window{
		Jumps: &Jumps{
			List: List[Jump]{},
		},
		cursors: cursors,
	}
}

// Jumps
type Jump struct {
	FilePath string
	Cursor   Cursor
}

type Jumps struct {
	List    List[Jump]
	current *Element[Jump]
}

func (j *Jumps) Push(b *Buffer, cur *Cursor) {
	// track only line jumps
	if j.List.Last() != nil {
		if j.List.Last().Value.FilePath == b.FilePath && j.List.Last().Value.Cursor.Line == cur.Line {
			return
		}
	}
	j.List.PushBack(Jump{
		FilePath: b.FilePath,
		Cursor:   *cur,
	})
	j.current = j.List.Last()
}

func (j *Jumps) JumpBack() {
	elem := j.List.Last()
	if elem == nil {
		return
	}

	if j.current != nil && j.current != elem {
		elem = j.current
	}

	if elem.Prev() == nil {
		return
	}

	item := elem.Prev().Value
	b := EditorInst.BufferFindByFilePath(item.FilePath, false)
	if b == nil {
		return
	}

	cur := CursorGet(EditorInst, b)
	cur.Line = item.Cursor.Line
	cur.Char = item.Cursor.Char
	cur.ScrollOffset = item.Cursor.ScrollOffset

	EditorInst.ActiveWindow().buf = b
	j.current = elem.Prev()
}

func (j *Jumps) JumpForward() {
	if j.current == nil {
		return
	}

	item := j.current.Next()
	if item == nil {
		return
	}

	b := EditorInst.BufferFindByFilePath(item.Value.FilePath, false)
	if b == nil {
		return
	}
	cur := CursorGet(EditorInst, b)
	cur.Line = item.Value.Cursor.Line
	cur.Char = item.Value.Cursor.Char
	cur.ScrollOffset = item.Value.Cursor.ScrollOffset
	EditorInst.ActiveWindow().buf = b
	j.current = item
}

