package mcwig

import "github.com/gdamore/tcell/v2"

type Viewport interface {
	Size() (width, height int)
}

type View interface {
	SetContent(x, y int, str string, st tcell.Style)
}

type Editor struct {
	Viewport     Viewport
	Keys         *KeyHandler
	Buffers      []*Buffer
	ActiveBuffer *Buffer
	ExitCh       chan int
}

func NewEditor(
	viewport Viewport,
	keys *KeyHandler,
) *Editor {
	return &Editor{
		Viewport:     viewport,
		Keys:         keys,
		Buffers:      []*Buffer{},
		ActiveBuffer: nil,
		ExitCh:       make(chan int),
	}
}

func (e *Editor) OpenFile(path string) {
	buf, err := BufferReadFile(path)
	if err != nil {
		panic(err)
	}
	e.Buffers = append(e.Buffers, buf)
	e.ActiveBuffer = buf
}

func (e *Editor) HandleInput(ev *tcell.EventKey) {
	buf := e.ActiveBuffer
	mode := buf.Mode
	e.Keys.HandleKey(e, ev, mode)
}
