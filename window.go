package mcwig

type Window struct {
	buf   *Buffer // active buffer
	Jumps *Jumps
}

// Jump to buffer and location. Records jump history.
func (win *Window) VisitBuffer(buf *Buffer, cursor ...Cursor) {
	if buf != nil {
		if win.buf != nil {
			win.Jumps.Push(win.buf)
		}

		if len(cursor) > 0 {
			buf.Cursor.Line = cursor[0].Line
			buf.Cursor.Char = cursor[0].Char
		}

		win.Jumps.Push(buf)
		win.buf = buf
	}
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

func CreateWindow() *Window {
	return &Window{
		Jumps: &Jumps{
			List: List[Jump]{},
		},
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

func (j *Jumps) Push(b *Buffer) {
	// track only line jumps
	if j.List.Last() != nil {
		if j.List.Last().Value.FilePath == b.FilePath && j.List.Last().Value.Cursor.Line == b.Cursor.Line {
			return
		}
	}

	j.List.PushBack(Jump{
		FilePath: b.FilePath,
		Cursor:   b.Cursor,
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
	b.Cursor = item.Cursor
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
	b.Cursor = item.Value.Cursor
	EditorInst.ActiveWindow().buf = b
	j.current = item
}
