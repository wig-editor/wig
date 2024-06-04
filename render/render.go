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

func (r *Renderer) Render() {
	r.screen.Clear()
	r.screen.Fill(' ', mcwig.Color("bg"))

	w, h := r.screen.Size()
	winWidth := w / len(r.e.Windows)

	// windows
	// TODO: rendering must be optimized.
	// - do not creste view every cycle. cache+reuse as much as possible.
	// - do not call Size(), instead use resize event
	for i, win := range r.e.Windows {
		winView := NewMView(r.screen, winWidth*i, 0, winWidth, h)
		ui.WindowRender(r.e, winView, win)
		ui.StatuslineRender(r.e, winView, win)
	}

	// widgets
	mainView := NewMView(r.screen, 0, 0, w, h)
	for _, c := range r.e.UiComponents {
		c.Render(mainView)
	}

	// ui.NotificationsRender(r.e, mainView)

	r.screen.Show()
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

func (t *mview) Resize(x, y, width, height int) {
	t.viewport.Resize(x, y, width, height)
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
