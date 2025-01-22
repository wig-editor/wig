package ui

import (
	"fmt"
	"strings"

	"github.com/firstrow/mcwig"
)

func StatuslineRender(
	e *mcwig.Editor,
	view mcwig.View,
	win *mcwig.Window,
) {
	buf := win.Buffer()

	w, h := view.Size()
	h -= 1

	st := mcwig.Color("ui.statusline.inactive")
	if win == e.ActiveWindow() {
		st = mcwig.Color("ui.statusline")
	}
	if buf.Mode() == mcwig.MODE_INSERT {
		st = mcwig.Color("ui.statusline.insert")
	}

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	leftSide := ""
	leftSide = fmt.Sprintf("%s %s", buf.Mode().String(), buf.GetName())

	if win.Buffer() == e.ActiveWindow().Buffer() && len(e.Message) > 0 {
		leftSide = e.Message
	}

	view.SetContent(2, h, leftSide, st)

	rightSide := fmt.Sprintf("%d:%d", buf.Cursor.Line+1, buf.Cursor.Char)

	if e.Keys.GetTimes() > 1 {
		rightSide = fmt.Sprintf("%d   %s", e.Keys.GetTimes(), rightSide)
	}

	view.SetContent(w-len(rightSide)-1, h, rightSide, st)
}
