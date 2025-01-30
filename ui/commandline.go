package ui

import (
	"fmt"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/gdamore/tcell/v2"
)

type uiCommandLine struct {
	e      *mcwig.Editor
	keymap *mcwig.KeyHandler
	chBuf  []rune
}

func CmdLineInit(ctx mcwig.Context) {
	cmdLine := &uiCommandLine{
		e:     ctx.Editor,
		chBuf: []rune{},
	}

	cmdLine.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			"Esc": func(ctx mcwig.Context) {
				ctx.Editor.PopUi()
			},
			"Tab": func(ctx mcwig.Context) {
				// todo autocomplete
			},
		},
	})
	cmdLine.keymap.Fallback(cmdLine.insertCh)

	ctx.Editor.PushUi(cmdLine)
}

func (u *uiCommandLine) insertCh(ctx mcwig.Context, ev *tcell.EventKey) {
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
			ctx.Editor.PopUi()
		}
		return
	}
	if ev.Key() == tcell.KeyEnter {
		cmd := strings.TrimSpace(string(u.chBuf))
		u.execute(cmd)
		ctx.Editor.PopUi()
		return
	}

	u.chBuf = append(u.chBuf, ev.Rune())
}

func (u *uiCommandLine) execute(cmd string) {
	ctx := u.e.NewContext()

	switch cmd {
	case "q":
		mcwig.CmdExit(ctx)
	case "q!":
		mcwig.CmdExit(ctx)
	case "w":
		mcwig.CmdSaveFile(ctx)
	case "bd":
		mcwig.CmdKillBuffer(ctx)
	}
}

func (u *uiCommandLine) Keymap() *mcwig.KeyHandler {
	return u.keymap
}

func (u *uiCommandLine) Render(view mcwig.View) {
	st := mcwig.Color("statusline")
	w, h := view.Size()
	h -= 1

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	msg := fmt.Sprintf(":%s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(0, h, msg, st)
}

func (u *uiCommandLine) Mode() mcwig.Mode {
	return mcwig.MODE_NORMAL
}
