package wig

import (
	"os"
	"path"

	"github.com/gdamore/tcell/v2"
)

type View interface {
	SetContent(x, y int, str string, st tcell.Style)
	Size() (width, height int)
	Resize(x, y, width, height int)
}

type RenderPlane int

const (
	PlaneWin    RenderPlane = 0
	PlaneEditor RenderPlane = 1
)

type UiComponent interface {
	Mode() Mode
	Keymap() *KeyHandler
	Render(view View)
	Plane() RenderPlane
}

type Context struct {
	Editor *Editor
	Buf    *Buffer
	Win    *Window
	Count  uint32
	Char   string
}

type AutocompleteFn func(Context) bool

var EditorInst *Editor

type Layout int

const (
	LayoutHorizontal Layout = 0
	LayoutVertical   Layout = 1
)

type Editor struct {
	View                View
	Keys                *KeyHandler
	Buffers             []*Buffer
	Windows             []*Window
	activeWindow        *Window
	UiComponents        []UiComponent
	ExitCh              chan int
	RedrawCh            chan int
	ScreenSyncCh        chan int
	Layout              Layout
	Yanks               List[yank]
	Projects            ProjectManager
	Message             string // display in echo area
	Lsp                 *LspManager
	Events              *EventsManager
	AutocompleteTrigger AutocompleteFn
	Snippets            *SnippetsManager
}

func NewEditor(
	view View,
	keys *KeyHandler,
) *Editor {
	windows := []*Window{CreateWindow(nil)}

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
		RedrawCh:     make(chan int, 10),
		ScreenSyncCh: make(chan int),
		Events:       NewEventsManager(),
		Snippets:     NewSnippetsManager(),
	}

	EditorInst.Lsp = NewLspManager(EditorInst)
	TreeSitterHighlighterGo(EditorInst)

	return EditorInst
}

func (e *Editor) OpenFile(path string) *Buffer {
	if fbuf := e.BufferFindByFilePath(path, false); fbuf != nil {
		return fbuf
	}

	var buf *Buffer
	buf, err := BufferReadFile(path)
	if err != nil {
		e.LogError(err)
		// create empty buffer
		buf = NewBuffer()
		buf.FilePath = path
	}

	e.Buffers = append(e.Buffers, buf)
	e.Lsp.DidOpen(buf)

	hl := TreeSitterHighlighterInitBuffer(e, buf)
	if hl != nil {
		buf.Highlighter = hl
	}

	return buf
}

func (e *Editor) NewContext() Context {
	return Context{
		Editor: e,
		Buf:    e.ActiveBuffer(),
		Win:    e.ActiveWindow(),
		Count:  0,
	}
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
	if len(e.Windows) > 1 {
		e.Windows[len(e.Windows)-1].ShowBuffer(b)
		return
	}
	win := CreateWindow(nil)
	win.buf = b
	e.Windows = append(e.Windows, win)
}

func (e *Editor) HandleInput(ev *tcell.EventKey) {
	var k *KeyHandler
	mode := e.ActiveBuffer().Mode()
	e.Message = ""

	if e.ActiveWindow().Buffer().KeyHandler != nil {
		k = e.ActiveWindow().Buffer().KeyHandler
	} else {
		k = e.Keys
	}

	if len(e.UiComponents) > 0 {
		comp := e.UiComponents[len(e.UiComponents)-1]
		k = comp.Keymap()
		mode = comp.Mode()
	}

	k.HandleKey(e, ev, mode)
}

func (e *Editor) LogError(err error, echo ...bool) {
	buf := e.BufferFindByFilePath("[Messages]", true)
	buf.Append("error: " + err.Error())
	if len(echo) > 0 && echo[0] == true {
		e.EchoMessage(err.Error())
	}
}

func (e *Editor) LogMessage(msg ...string) {
	for _, m := range msg {
		buf := e.BufferFindByFilePath("[Messages]", true)
		buf.Append(m)
	}
}

func (e *Editor) RuntimeDir(elems ...string) string {
	p := []string{os.Getenv("HOME"), ".config", "wig"}
	elems = append(p, elems...)
	return path.Join(elems...)
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

