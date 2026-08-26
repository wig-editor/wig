package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/wig"
)

// FloatingWindowOpts configures a FloatingWindow's appearance and behavior.
type FloatingWindowOpts struct {
	// Width is the fraction of the screen width (0.0–1.0).
	Width float64
	// Height is the fraction of the screen height (0.0–1.0).
	Height float64
	// Title displayed in the top border.
	Title string
}

// OffsetView wraps a View and translates relative coordinates to absolute
// screen coordinates, clipping to a sub-rectangle. This lets us reuse
// WindowRender (which expects a full View) inside a bordered region.
type OffsetView struct {
	inner            wig.View
	offsetX, offsetY int
	width, height    int
}

func (ov *OffsetView) SetContent(x, y int, str string, st tcell.Style) {
	if x < 0 || x >= ov.width || y < 0 || y >= ov.height {
		return
	}
	// Clip the string so it doesn't overflow past the right edge of
	// the floating window interior and draw over the border.
	maxChars := ov.width - x
	if maxChars <= 0 {
		return
	}
	runes := []rune(str)
	if len(runes) > maxChars {
		str = string(runes[:maxChars])
	}
	ov.inner.SetContent(ov.offsetX+x, ov.offsetY+y, str, st)
}

func (ov *OffsetView) Size() (int, int) {
	return ov.width, ov.height
}

func (ov *OffsetView) Resize(x, y, width, height int) {
	ov.offsetX = x
	ov.offsetY = y
	ov.width = width
	ov.height = height
}

// FloatingWindow displays a Buffer in a centered, bordered overlay.
// It implements wig.UiComponent so it integrates with the existing
// UI stack (input routing, rendering).
type FloatingWindow struct {
	editor     *wig.Editor
	buf        *wig.Buffer
	win        *wig.Window
	opts       FloatingWindowOpts
	keyHandler *wig.KeyHandler
}

// NewFloatingWindow creates a floating window for the given buffer.
// If Width or Height are ≤ 0, the editor's config defaults are used.
func NewFloatingWindow(editor *wig.Editor, buf *wig.Buffer, opts FloatingWindowOpts) *FloatingWindow {
	if opts.Width <= 0 {
		opts.Width = editor.Config.FloatingWindowWidth
		if opts.Width <= 0 {
			opts.Width = 0.8
		}
	}
	if opts.Height <= 0 {
		opts.Height = editor.Config.FloatingWindowHeight
		if opts.Height <= 0 {
			opts.Height = 0.8
		}
	}

	fw := &FloatingWindow{
		editor: editor,
		buf:    buf,
		win:    wig.CreateWindow(editor.ActiveWindow()),
		opts:   opts,
	}
	fw.win.ShowBuffer(buf)

	// Ensure a cursor exists for this buffer in the floating window.
	if wig.WindowCursorGet(fw.win, buf) == nil {
		wig.WindowCursorSet(fw.win, buf, &wig.Cursor{Line: 0, Char: 0})
	}

	kh := wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc":    func(ctx wig.Context) { fw.Close() },
			"ctrl+c": func(ctx wig.Context) { fw.Close() },
		},
	})
	kh.EnableTimes = true
	kh.SetContextFn(func(editor *wig.Editor) wig.Context {
		return wig.Context{
			Editor: editor,
			Buf:    fw.buf,
			Win:    fw.win,
		}
	})
	fw.keyHandler = kh

	return fw
}

// Close removes the floating window from the UI stack.
func (fw *FloatingWindow) Close() {
	if len(fw.editor.UiComponents) > 0 && fw.editor.UiComponents[len(fw.editor.UiComponents)-1] == fw {
		fw.editor.PopUi()
	} else {
		fw.editor.PopUiComponent(fw)
	}
	fw.editor.Redraw()
}

// Buffer returns the buffer displayed in this floating window.
func (fw *FloatingWindow) Buffer() *wig.Buffer { return fw.buf }

// Window returns the internal window used for cursor management.
func (fw *FloatingWindow) Window() *wig.Window { return fw.win }

// SetKeyHandler replaces the floating window's key handler. The context
// function is automatically configured to point at the floating window's
// buffer and internal window.
func (fw *FloatingWindow) SetKeyHandler(kh *wig.KeyHandler) {
	kh.SetContextFn(func(editor *wig.Editor) wig.Context {
		return wig.Context{
			Editor: editor,
			Buf:    fw.buf,
			Win:    fw.win,
		}
	})
	fw.keyHandler = kh
}

// SetCursor positions the cursor within the floating window.
func (fw *FloatingWindow) SetCursor(line, char int) {
	cur := wig.WindowCursorGet(fw.win, fw.buf)
	if cur != nil {
		cur.Line = line
		cur.Char = char
	}
}

// SetCursorToEnd moves the cursor to the last line of the buffer.
func (fw *FloatingWindow) SetCursorToEnd() {
	if fw.buf.Lines.Len > 0 {
		fw.SetCursor(fw.buf.Lines.Len-1, 0)
	}
}

// SetTitle updates the title shown in the top border.
func (fw *FloatingWindow) SetTitle(title string) {
	fw.opts.Title = title
}

// UiComponent interface

func (fw *FloatingWindow) Plane() wig.RenderPlane  { return wig.PlaneEditor }
func (fw *FloatingWindow) Mode() wig.Mode          { return fw.buf.Mode() }
func (fw *FloatingWindow) Keymap() *wig.KeyHandler { return fw.keyHandler }

// Render draws the floating window: background, border, title, and
// the buffer content rendered via WindowRender into a clipped sub-view.
func (fw *FloatingWindow) Render(view wig.View) {
	vw, vh := view.Size()

	width := int(float64(vw) * fw.opts.Width)
	height := int(float64(vh) * fw.opts.Height)

	if width < 10 {
		width = 10
	}
	if height < 3 {
		height = 3
	}
	if width > vw {
		width = vw
	}
	if height > vh {
		height = vh
	}

	x := (vw - width) / 2
	y := (vh - height) / 2

	borderStyle := wig.Color("ui.window.border")
	bgStyle := wig.Color("ui.background")

	// Fill background
	for row := 0; row < height; row++ {
		view.SetContent(x, y+row, strings.Repeat(" ", width), bgStyle)
	}

	// Draw border (excluding corners)
	for col := 1; col < width-1; col++ {
		view.SetContent(x+col, y, "─", borderStyle)
		view.SetContent(x+col, y+height-1, "─", borderStyle)
	}
	for row := 1; row < height-1; row++ {
		view.SetContent(x, y+row, "│", borderStyle)
		view.SetContent(x+width-1, y+row, "│", borderStyle)
	}

	// Corners
	view.SetContent(x, y, "┌", borderStyle)
	view.SetContent(x+width-1, y, "┐", borderStyle)
	view.SetContent(x, y+height-1, "└", borderStyle)
	view.SetContent(x+width-1, y+height-1, "┘", borderStyle)

	// Title in top border
	if fw.opts.Title != "" {
		titleStr := " " + fw.opts.Title + " "
		maxLen := width - 4
		if maxLen < 1 {
			maxLen = 1
		}
		if len([]rune(titleStr)) > maxLen {
			titleStr = string([]rune(titleStr)[:maxLen])
		}
		view.SetContent(x+2, y, titleStr, borderStyle)
	}

	// Render buffer content inside the border
	interiorW := width - 2
	interiorH := height - 2
	if interiorW > 0 && interiorH > 0 {
		offsetView := &OffsetView{
			inner:   view,
			offsetX: x + 1,
			offsetY: y + 1,
			width:   interiorW,
			height:  interiorH,
		}
		WindowRender(fw.editor, offsetView, fw.win)
	}
}
