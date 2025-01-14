package mcwig

type Window struct {
	buf   *Buffer // active buffer
	Jumps *Jumps
}

func (win *Window) SetBuffer(buf *Buffer) {
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
	BufferName string
	FilePath   string
	Cursor     Cursor
}

type Jumps struct {
	List List[Jump]
}

func (j *Jumps) Push(b *Buffer) {
	j.List.PushBack(Jump{
		BufferName: b.GetName(),
		FilePath:   b.FilePath,
		Cursor:     b.Cursor,
	})
}
