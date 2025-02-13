package ui

import (
	"github.com/firstrow/mcwig"
	"go.lsp.dev/protocol"
)

type AutocompleteWidget struct {
	ctx        mcwig.Context
	triggerPos mcwig.Cursor
	keymap     *mcwig.KeyHandler
	pos        mcwig.Position
	items      protocol.CompletionList
	activeItem int
}

func (u *AutocompleteWidget) Plane() mcwig.RenderPlane {
	return mcwig.PlaneEditor
}

func AutocompleteInit(ctx mcwig.Context, pos mcwig.Position, items protocol.CompletionList) *AutocompleteWidget {
	if len(items.Items) == 0 {
		return nil
	}

	widget := &AutocompleteWidget{
		ctx:        ctx,
		pos:        pos,
		items:      items,
		activeItem: 0,
	}

	widget.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_INSERT: mcwig.KeyMap{
			"Esc": func(ctx mcwig.Context) {
				ctx.Editor.PopUi()
			},
			"Tab": func(ctx mcwig.Context) {
				widget.activeItem++
			},
			"Enter": widget.selectItem,
		},
	})

	// watch text change event for filter

	ctx.Editor.PushUi(widget)
	return widget
}

func (w *AutocompleteWidget) Mode() mcwig.Mode {
	return mcwig.MODE_INSERT
}

func (w *AutocompleteWidget) Keymap() *mcwig.KeyHandler {
	return w.keymap
}

func (w *AutocompleteWidget) selectItem(ctx *mcwig.Context) {

}

func (w *AutocompleteWidget) Render(view mcwig.View) {
	x := w.pos.Char + 2
	y := w.pos.Line - w.ctx.Buf.ScrollOffset + 1

	maxItems := min(10, len(w.items.Items)-1)

	drawBoxNoBorder(view, w.pos.Char, y, 50, maxItems, mcwig.Color("ui.popup"))

	for i, row := range w.items.Items {
		view.SetContent(x, y, row.Label, mcwig.Color("ui.popup"))
		if i >= maxItems {
			return
		}
		y++
	}
}

