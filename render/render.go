package render

import (
	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
	"github.com/gdamore/tcell/v2/views"
	"github.com/mattn/go-runewidth"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/ui"
)

type Renderer struct {
	e      *mcwig.Editor
	screen tcell.Screen
}

func New(e *mcwig.Editor, screen tcell.Screen) *Renderer {
	r := &Renderer{
		e:      e,
		screen: screen,
	}
	return r
}

func (r *Renderer) SetContent(x, y int, str string, st tcell.Style) {
	xx := x
	for _, ch := range str {
		var comb []rune
		w := runewidth.RuneWidth(ch)
		if w == 0 {
			comb = []rune{ch}
			ch = ' '
			w = 1
		}

		r.screen.SetContent(xx, y, ch, comb, st)
		xx += w
	}
}

func (r *Renderer) Render() {
	r.screen.Clear()
	r.screen.Fill(' ', mcwig.Color("bg"))

	// get all windows
	// render window as a plane

	buf := r.e.ActiveBuffer()
	if buf == nil {
		return
	}

	currentLine := buf.Lines.First()
	offset := buf.ScrollOffset
	lineNum := 0
	y := 0

	for currentLine != nil {
		if lineNum >= offset {
			// render each character in the line separately
			x := 0

			// render cursor on empty line
			if len(currentLine.Value) == 0 && lineNum == buf.Cursor.Line {
				r.SetContent(x, y, " ", mcwig.Color("cursor"))
			}

			for i := 0; i < len(currentLine.Value); i++ {

				// render selection
				textStyle := mcwig.Color("default")
				if buf.Selection != nil {
					if mcwig.SelectionCursorInRange(buf.Selection, mcwig.Cursor{Line: lineNum, Char: i}) {
						textStyle = mcwig.Color("statusline")
					}
				}

				ch := getRenderChar(currentLine.Value[i])
				r.SetContent(x, y, string(ch), textStyle)

				// render cursor
				if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
					r.SetContent(x, y, string(ch[0]), mcwig.Color("cursor"))
				}

				x += len(ch)
			}

			// render cursor after the end of the line in insert mode
			if lineNum == buf.Cursor.Line && buf.Cursor.Char >= len(currentLine.Value) {
				r.SetContent(x, y, " ", mcwig.Color("cursor"))
			}

			y++
		}

		currentLine = currentLine.Next()
		lineNum++
	}

	tview := NewMView(r.screen, 10, 3, 100, 100)

	ui.StatuslineRender(r.e, tview)

	for _, c := range r.e.UiComponents {
		c.Render(tview)
	}

	r.screen.Show()
}

type mview struct {
	viewport *views.ViewPort
}

func NewMView(view views.View, x, y, width, height int) *mview {
	return &mview{
		viewport: views.NewViewPort(view, 10, 10, 100, 100),
	}
}

func (t *mview) Size() (int, int) {
	return t.viewport.Size()
}

func (t *mview) SetContent(x, y int, str string, st tcell.Style) {
	xx := x
	for _, ch := range str {
		var comb []rune
		w := runewidth.RuneWidth(ch)
		if w == 0 {
			comb = []rune{ch}
			ch = ' '
			w = 1
		}

		t.viewport.SetContent(xx, y, ch, comb, st)
		xx += w
	}
}

func getRenderChar(ch rune) string {
	if ch == '\t' {
		return "    "
	}
	return string(ch)
}
