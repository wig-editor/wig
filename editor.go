package mcwig

import "github.com/gdamore/tcell/v2"

type View interface {
	SetContent(x, y int, str string, st tcell.Style)
	Size() (width, height int)
}

type UiComponent interface {
	Mode() Mode
	Keymap() *KeyHandler
	Render(view View)
}

type Editor struct {
	View         View
	Keys         *KeyHandler
	Buffers      []*Buffer
	Windows      []*Window
	UiComponents []UiComponent
	ExitCh       chan int
	RedrawCh     chan int

	activeWindow *Window
}

func NewEditor(
	view View,
	keys *KeyHandler,
) *Editor {
	windows := []*Window{{}}

	return &Editor{
		View:         view,
		Keys:         keys,
		Buffers:      []*Buffer{},
		Windows:      windows,
		activeWindow: windows[0],
		ExitCh:       make(chan int),
		RedrawCh:     make(chan int, 10),
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

func (e *Editor) Redraw() {
	e.RedrawCh <- 1
}
