package ui

import "github.com/firstrow/mcwig"

type uiCommandLine struct {
	e      *mcwig.Editor
	keymap *mcwig.KeyHandler
}

func CommandLineInit(e *mcwig.Editor) {
	cmdLine := &uiCommandLine{e: e}
	cmdLine.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			"Tab": func(e *mcwig.Editor) {
				// panic("WORKS")
			},
			"Esc": func(e *mcwig.Editor) {
				e.PopUi()
			},
		},
	})

	e.PushUi(cmdLine)
}

func (u *uiCommandLine) Mode() mcwig.Mode {
	return mcwig.MODE_NORMAL
}

func (u *uiCommandLine) Keymap() *mcwig.KeyHandler {
	return u.keymap
}

func (u *uiCommandLine) Render(view mcwig.View, viewport mcwig.Viewport) {
	view.SetContent(0, 0, "YAY", mcwig.Color("cursor"))
}
