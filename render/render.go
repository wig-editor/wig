package render

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
	"github.com/gdamore/tcell/v2/views"
	"github.com/mattn/go-runewidth"

	"github.com/firstrow/wig"
	"github.com/firstrow/wig/ui"
)

type Renderer struct {
	rw     sync.Mutex
	e      *wig.Editor
	screen tcell.Screen
}

func New(e *wig.Editor, screen tcell.Screen) *Renderer {
	r := &Renderer{
		e:      e,
		screen: screen,
	}
	return r
}

// TODO: rendering must be optimized.
func (r *Renderer) Render() {
	// TODO: schedule render
	r.rw.Lock()
	defer r.rw.Unlock()

	r.screen.Fill(' ', wig.Color("ui.background"))

	w, h := r.screen.Size()

	var winW int
	// var winH int
	if r.e.Layout == wig.LayoutVertical {
		winW = w / len(r.e.Windows)
		// winH = h
	} else {
		winW = w
		// winH = h / len(r.e.Windows)
	}

	// windows
	var winView *mview
	var activeWinView *mview
	for i, win := range r.e.Windows {
		// TODO: for now Vertical only. I never use horizontal splits
		// if r.e.Layout == wig.LayoutVertical {
		x := winW * i
		if i > 0 {
			st := wig.Color("ui.virtual.indent-guide")
			for i := 0; i <= h; i++ {
				r.SetContent(x, i, string(tcell.RuneVLine), st)
			}
			x += 1
		}
		winView = NewMView(r.screen, x, 0, winW, h)

		if win == r.e.ActiveWindow() {
			activeWinView = winView
		}

		ui.WindowRender(r.e, winView, win)
		ui.StatuslineRender(r.e, winView, win)
	}

	// widgets: pickers, etc...
	mainView := NewMView(r.screen, 0, 0, w, h)
	for _, c := range r.e.UiComponents {
		switch c.Plane() {
		case wig.PlaneWin:
			c.Render(activeWinView)
		default:
			c.Render(mainView)
		}
	}

	// ui.NotificationsRender(r.e, mainView)
	r.screen.Show()
}

func (r *Renderer) SetContent(x, y int, str string, st tcell.Style) {
	for _, ch := range str {
		var comb []rune
		w := runewidth.RuneWidth(ch)
		if w == 0 {
			comb = []rune{ch}
			ch = ' '
			w = 1
		}

		r.screen.SetContent(x, y, ch, comb, st)
		x += w
	}
}

func (r *Renderer) RenderMetrics(info map[string]time.Duration) {
	y := 0
	for k, v := range info {
		r.SetContent(50, y, fmt.Sprintf("%s: %v", k, v), tcell.StyleDefault)
		y++
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

func (t *mview) Resize(x, y, width, height int) {
	t.viewport.Resize(x, y, width, height)
}

func (t *mview) SetContent(x, y int, str string, st tcell.Style) {
	for _, ch := range str {
		var comb []rune
		w := runewidth.RuneWidth(ch)
		if w == 0 {
			comb = []rune{ch}
			ch = ' '
			w = 1
		}

		t.viewport.SetContent(x, y, ch, comb, st)
		x += w
	}
}

