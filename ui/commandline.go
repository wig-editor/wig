package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/mcwig"
)

type uiCommandLine struct {
	e      *mcwig.Editor
	keymap *mcwig.KeyHandler

	chBuf []rune
}

func CommandLineInit(e *mcwig.Editor) {
	cmdLine := &uiCommandLine{
		e:     e,
		chBuf: []rune{},
	}
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
	cmdLine.keymap.Fallback(cmdLine.insertCh)

	e.PushUi(cmdLine)
}

func (u *uiCommandLine) insertCh(e *mcwig.Editor, ev *tcell.EventKey) {
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		return
	}

	if ev.Modifiers()&tcell.ModAlt != 0 {
		return
	}

	if ev.Modifiers()&tcell.ModMeta != 0 {
		return
	}

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		if len(u.chBuf) > 0 {
			u.chBuf = u.chBuf[:len(u.chBuf)-1]
		} else {
			e.PopUi()
		}
		return
	}
	if ev.Key() == tcell.KeyEnter {
		e.PopUi()
		return
	}

	u.chBuf = append(u.chBuf, ev.Rune())
}

func (u *uiCommandLine) Mode() mcwig.Mode {
	return mcwig.MODE_NORMAL
}

func (u *uiCommandLine) Keymap() *mcwig.KeyHandler {
	return u.keymap
}

func (u *uiCommandLine) Render(view mcwig.View, viewport mcwig.Viewport) {
	st := mcwig.Color("statusline")
	w, h := viewport.Size()
	h -= 1

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	msg := fmt.Sprintf(":%s", string(u.chBuf))
	view.SetContent(0, h, msg, st)
}
