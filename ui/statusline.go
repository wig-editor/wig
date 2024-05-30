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
	buf := win.Buffer
	w, h := view.Size()
	h = h - 1
	st := mcwig.Color("statusline")

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	leftSide := fmt.Sprintf("%s %s", buf.Mode.String(), buf.FilePath)
	view.SetContent(2, h, leftSide, st)

	rightSide := fmt.Sprintf("%d:%d", buf.Cursor.Line+1, buf.Cursor.Char)
	view.SetContent(w-len(rightSide)-1, h, rightSide, st)
}
