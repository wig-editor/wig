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

var EditorInst *Editor

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
	Lsp          *LspManager

	activeWindow *Window
}

func NewEditor(
	view View,
	keys *KeyHandler,
) *Editor {
	windows := []*Window{CreateWindow()}

	EditorInst = &Editor{
		View:         view,
		Keys:         keys,
		Buffers:      make([]*Buffer, 0, 32),
		Yanks:        List[yank]{},
		Windows:      windows,
		activeWindow: windows[0],
		Layout:       LayoutVertical,
		Projects:     NewProjectManager(),
		ExitCh:       make(chan int),
		RedrawCh:     make(chan int),
		ScreenSyncCh: make(chan int),
	}
	EditorInst.Lsp = NewLspManager(EditorInst)

	return EditorInst
}

func (e *Editor) OpenFile(path string) *Buffer {
	if fbuf := e.BufferFindByFilePath(path, false); fbuf != nil {
		return fbuf
	}

	buf, err := BufferReadFile(path)
	if err != nil {
		e.LogError(err)
		return nil
	}
	e.Buffers = append(e.Buffers, buf)

	e.Lsp.DidOpen(buf)

	HighlighterInitBuffer(e, buf)

	return buf
}

// Find or create new buffer by its full file path
func (e *Editor) BufferFindByFilePath(fp string, create bool) *Buffer {
	for _, b := range e.Buffers {
		if b.FilePath == fp {
			return b
		}
	}

	if !create {
		return nil
	}

	b := NewBuffer()
	b.FilePath = fp
	b.Lines = List[Line]{}
	e.Buffers = append(e.Buffers, b)

	return b
}

// Returns active window buffer
func (e *Editor) ActiveBuffer() *Buffer {
	if len(e.Buffers) == 0 {
		buf := NewBuffer()
		buf.FilePath = "[No Name]"
		e.Buffers = append(e.Buffers, buf)
		e.ActiveWindow().ShowBuffer(buf)
	}

	return e.ActiveWindow().Buffer()
}

func (e *Editor) ActiveWindow() *Window {
	return e.activeWindow
}

func (e *Editor) SetActiveWindow(w *Window) {
	e.activeWindow = w
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
	for _, win := range e.Windows {
		if win.Buffer() == b {
			return
		}
	}
	if len(e.Windows) >= 1 {
		e.Windows[len(e.Windows)-1].ShowBuffer(b)
		return
	}
	e.Windows = append(e.Windows, &Window{buf: b})
}

func (e *Editor) HandleInput(ev *tcell.EventKey) {
	mode := e.ActiveBuffer().Mode()
	h := e.Keys.HandleKey
	e.Message = ""

	if len(e.UiComponents) > 0 {
		comp := e.UiComponents[len(e.UiComponents)-1]
		h = comp.Keymap().HandleKey
		mode = comp.Mode()
	}

	h(e, ev, mode)
}

func (e *Editor) LogError(err error, echo ...bool) {
	buf := e.BufferFindByFilePath("[Messages]", true)
	buf.Append("error: " + err.Error())
	if len(echo) > 0 {
		e.EchoMessage(err.Error())
	}
}

func (e *Editor) LogMessage(msg string) {
	buf := e.BufferFindByFilePath("[Messages]", true)
	buf.Append(msg)
}

func (e *Editor) EchoMessage(msg string) {
	buf := e.BufferFindByFilePath("[Messages]", true)
	buf.Append(msg)
	e.Message = msg
}

func (e *Editor) Redraw() {
	e.RedrawCh <- 1
}

func (e *Editor) ScreenSync() {
	e.ScreenSyncCh <- 1
}
