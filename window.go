package mcwig

type Window struct {
	buf   *Buffer // active buffer
	Jumps *Jumps
}

func (win *Window) SetBuffer(buf *Buffer) {
	if buf != nil {
		win.Jumps.Push(buf)
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
	EditorInst.ActiveWindow().SetBuffer(b)
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
	EditorInst.ActiveWindow().SetBuffer(b)
	j.current = item
}
