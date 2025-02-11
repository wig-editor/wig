package ui

import "github.com/firstrow/mcwig"

type AutocompleteWidget struct {
	triggerPos mcwig.Cursor
	keymap     *mcwig.KeyHandler
}

func AutocompleteInit(e *mcwig.Editor) *AutocompleteWidget {
	widget := &AutocompleteWidget{}

	widget.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_INSERT: mcwig.KeyMap{
			"Esc": func(ctx mcwig.Context) {
				ctx.Editor.PopUi()
			},
		},
	})

	// watch text change event for filter

	e.PushUi(widget)
	return widget
}

func (w *AutocompleteWidget) Mode() mcwig.Mode {
	return mcwig.MODE_INSERT
}
func (w *AutocompleteWidget) Keymap() *mcwig.KeyHandler {
	return w.keymap
}
func (w *AutocompleteWidget) Render(view mcwig.View) {
	drawBoxNoBorder(view, 10, 13, 50, 15, mcwig.Color("ui.popup"))
}

