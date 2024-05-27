package mcwig

import "github.com/gdamore/tcell/v2"

type Viewport interface {
	Size() (width, height int)
}

type View interface {
	SetContent(x, y int, str string, st tcell.Style)
}

type UiComponent interface {
	Mode() Mode
	Keymap() *KeyHandler
	Render(view View, viewport Viewport)
}

type Editor struct {
	Viewport     Viewport
	Keys         *KeyHandler
	Buffers      []*Buffer
	ActiveBuffer *Buffer
	UiComponents []UiComponent
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

func (e *Editor) PushUi(c UiComponent) {
	e.UiComponents = append(e.UiComponents, c)
}

func (e *Editor) PopUi() {
	if len(e.UiComponents) > 0 {
		e.UiComponents = e.UiComponents[:len(e.UiComponents)-1]
	}
}

func (e *Editor) HandleInput(ev *tcell.EventKey) {
	mode := e.ActiveBuffer.Mode
	h := e.Keys.HandleKey

	if len(e.UiComponents) > 0 {
		h = e.UiComponents[len(e.UiComponents)-1].Keymap().HandleKey
	}

	h(e, ev, mode)
}
