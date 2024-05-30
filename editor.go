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
	Windows      []*Window
	UiComponents []UiComponent
	ExitCh       chan int

	activeWindow *Window
}

func NewEditor(
	viewport Viewport,
	keys *KeyHandler,
) *Editor {
	windows := []*Window{
		&Window{},
	}

	return &Editor{
		Viewport:     viewport,
		Keys:         keys,
		Buffers:      []*Buffer{},
		Windows:      windows,
		activeWindow: windows[0],
		ExitCh:       make(chan int),
	}
}

func (e *Editor) OpenFile(path string) {
	buf, err := BufferReadFile(path)
	if err != nil {
		panic(err)
	}
	e.Buffers = append(e.Buffers, buf)
	e.activeWindow.Buffer = buf
}

func (e *Editor) ActiveBuffer() *Buffer {
	if len(e.Buffers) == 0 {
		buf := NewBuffer()
		buf.Name = "[No Name]"
		e.Buffers = append(e.Buffers, buf)
		e.activeWindow.Buffer = buf
	}

	return e.activeWindow.Buffer
}

func (e *Editor) ActiveWindow() *Window {
	return e.activeWindow
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
	mode := e.ActiveBuffer().Mode
	h := e.Keys.HandleKey

	if len(e.UiComponents) > 0 {
		h = e.UiComponents[len(e.UiComponents)-1].Keymap().HandleKey
	}

	h(e, ev, mode)
}
