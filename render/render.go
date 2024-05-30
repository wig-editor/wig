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

	w, h := r.screen.Size()

	tview := NewMView(r.screen, 0, 0, w/2, h)

	ui.WindowRender(r.e, tview, r.e.Windows[0])
	ui.StatuslineRender(r.e, tview, r.e.Windows[0])

	for _, c := range r.e.UiComponents {
		c.Render(tview)
	}

	r.screen.Show()
}

type mview struct {
	viewport *views.ViewPort
}

func NewMView(view views.View, x, y, w, h int) *mview {
	return &mview{
		viewport: views.NewViewPort(view, x, y, w, h),
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
