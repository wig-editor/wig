package ui

import (
	"github.com/firstrow/mcwig"
)

func NotificationsRender(e *mcwig.Editor, view mcwig.View) {
	return
	vw, vh := view.Size()

	x := vw - 53
	y := vh - 5
	w := 50
	h := 2

	drawBox2(view, x, y, w, h, mcwig.Color("statusline"))
	view.SetContent(x+1, y+1, truncate("sdf", 48), mcwig.Color("statusline"))
}
