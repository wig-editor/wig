package ui

import (
	"fmt"
	"strings"

	"github.com/firstrow/wig"
)

func StatuslineRender(
	e *wig.Editor,
	view wig.View,
	win *wig.Window,
) {
	buf := win.Buffer()
	if buf == nil {
		return
	}

	w, h := view.Size()
	h -= 1

	st := wig.Color("ui.statusline.inactive")

	if win == e.ActiveWindow() {
		st = wig.Color("ui.statusline")

		if buf.Mode() == wig.MODE_INSERT {
			st = wig.Color("ui.statusline.insert")
		}
	}

	bg := strings.Repeat(" ", w)
	if h >= 0 {
		view.SetContent(0, h, bg, st)
	}

	macroStatus := ""
	if e.Keys.Macros.Recording() {
		macroStatus = "recording @" + e.Keys.Macros.Register
	}

	leftSide := fmt.Sprintf("%s %s %s ", buf.Mode().String(), buf.GetName(), macroStatus)

	if win.Buffer() == e.ActiveWindow().Buffer() && len(e.Message) > 0 {
		leftSide = e.Message
	}

	if h >= 0 {
		view.SetContent(2, h, leftSide, st)
	}

	cur := wig.CursorGet(e, buf)
	rightSide := fmt.Sprintf("[workspace: %d] %d:%d", e.ActiveWorkspace, cur.Line+1, cur.Char)

	if e.Keys.GetCount() > 1 {
		rightSide = fmt.Sprintf("%d   %s", e.Keys.GetCount(), rightSide)
	}

	if w-len(rightSide)-1 >= 0 && h >= 0 {
		view.SetContent(w-len(rightSide)-1, h, rightSide, st)
	}
}
