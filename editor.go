package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

type View interface {
	SetContent(x, y int, str string, st tcell.Style)
	Size() (width, height int)
	Resize(x, y, width, height int)
}

type UiComponent interface {
	Mode() Mode
	Keymap() *KeyHandler
	Render(view View)
}

type Layout int

const (
	LayoutHorizontal Layout = 0
	LayoutVertical   Layout = 1
)

type Editor struct {
	View         View
	Keys         *KeyHandler
	Buffers      []*Buffer
	Windows      []*Window
	UiComponents []UiComponent
	ExitCh       chan int
	RedrawCh     chan int
	ScreenSyncCh chan int
	Layout       Layout
	Yanks        List[yank]
	Projects     ProjectManager
	Message      string // display in echo area

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
		Yanks:        List[yank]{},
		Windows:      windows,
		activeWindow: windows[0],
		ExitCh:       make(chan int),
		RedrawCh:     make(chan int, 10),
		ScreenSyncCh: make(chan int),
		Layout:       LayoutVertical,
		Projects:     NewProjectManager(),
	}
}

func (e *Editor) OpenFile(path string) {
	buf, err := BufferReadFile(path)
	if err != nil {
		e.LogError(err)
		return
	}
	e.Buffers = append(e.Buffers, buf)
	e.activeWindow.Buffer = buf
}

func (e *Editor) ActiveBuffer() *Buffer {
	if len(e.Buffers) == 0 {
		buf := NewBuffer()
		buf.FilePath = "[No Name]"
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

func (e *Editor) EnsureBufferIsVisible(b *Buffer) {
	found := false
	for _, win := range e.Windows {
		if win.Buffer == b {
			found = true
		}
	}
	if found {
		return
	}
	if len(e.Windows) > 1 {
		e.Windows[len(e.Windows)-1].Buffer = b
		return
	}
	e.Windows = append(e.Windows, &Window{Buffer: b})
}

func (e *Editor) HandleInput(ev *tcell.EventKey) {
	mode := e.ActiveBuffer().Mode
	h := e.Keys.HandleKey
	e.Message = ""

	if len(e.UiComponents) > 0 {
		h = e.UiComponents[len(e.UiComponents)-1].Keymap().HandleKey
	}

	h(e, ev, mode)
}

func (e *Editor) LogError(err error) {
	buf := e.BufferFindByFilePath("[Messages]")
	buf.Append("error: " + err.Error())
}

func (e *Editor) LogMessage(msg string) {
	buf := e.BufferFindByFilePath("[Messages]")
	buf.Append(msg)
}

func (e *Editor) EchoMessage(msg string) {
	buf := e.BufferFindByFilePath("[Messages]")
	buf.Append(msg)
	e.Message = msg
}

func (e *Editor) Redraw() {
	e.RedrawCh <- 1
}

func (e *Editor) ScreenSync() {
	e.ScreenSyncCh <- 1
}
