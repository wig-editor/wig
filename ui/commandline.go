package ui

import (
	"fmt"
	"strings"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

type uiCommandLine struct {
	e      *wig.Editor
	keymap *wig.KeyHandler
	chBuf  []rune
}

func (u *uiCommandLine) Plane() wig.RenderPlane {
	return wig.PlaneEditor
}

func CmdLineInit(ctx wig.Context) {
	cmdLine := &uiCommandLine{
		e:     ctx.Editor,
		chBuf: []rune{},
	}

	cmdLine.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc": func(ctx wig.Context) {
				ctx.Editor.PopUi()
			},
			"Tab": func(ctx wig.Context) {
				// todo autocomplete
			},
		},
	})
	cmdLine.keymap.Fallback(cmdLine.insertCh)

	ctx.Editor.PushUi(cmdLine)
}

func (u *uiCommandLine) insertCh(ctx wig.Context, ev *tcell.EventKey) {
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
		wig.CmdExit(ctx)
	case "q!":
		wig.CmdExit(ctx)
	case "w":
		wig.CmdSaveFile(ctx)
	case "bd":
		wig.CmdKillBuffer(ctx)
	}
}

func (u *uiCommandLine) Keymap() *wig.KeyHandler {
	return u.keymap
}

func (u *uiCommandLine) Render(view wig.View) {
	st := wig.Color("statusline")
	w, h := view.Size()
	h -= 1

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	msg := fmt.Sprintf(":%s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(0, h, msg, st)
}

func (u *uiCommandLine) Mode() wig.Mode {
	return wig.MODE_NORMAL
}

