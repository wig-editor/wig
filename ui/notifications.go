package ui

import (
	"github.com/firstrow/wig"
)

func NotificationsRender(e *wig.Editor, view wig.View) {
	vw, vh := view.Size()

	x := vw - 53
	y := vh - 5
	w := 50
	h := 2

	drawBox2(view, x, y, w, h, wig.Color("statusline"))
	view.SetContent(x+1, y+1, truncate("sdf", 48), wig.Color("statusline"))
}
