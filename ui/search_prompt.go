package ui

import (
	"fmt"
	"strings"

	"github.com/firstrow/mcwig"
	"github.com/gdamore/tcell/v2"
)

type uiSearchPrompt struct {
	e      *mcwig.Editor
	keymap *mcwig.KeyHandler

	chBuf []rune
}

func CmdSearchPromptInit(ctx mcwig.Context) {
	cmdLine := &uiSearchPrompt{
		e:     ctx.Editor,
		chBuf: []rune{},
	}

	cmdLine.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			"Esc": func(ctx mcwig.Context) {
				ctx.Editor.PopUi()
			},
		},
	})
	cmdLine.keymap.Fallback(cmdLine.insertCh)
	ctx.Editor.PushUi(cmdLine)
}

func (u *uiSearchPrompt) insertCh(ctx mcwig.Context, ev *tcell.EventKey) {
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

func (u *uiSearchPrompt) execute(cmd string) {
	pat := strings.TrimSpace(cmd)
	mcwig.LastSearchPattern = pat
	mcwig.SearchNext(u.e.NewContext(), pat)
}

func (u *uiSearchPrompt) Keymap() *mcwig.KeyHandler {
	return u.keymap
}

func (u *uiSearchPrompt) Render(view mcwig.View) {
	st := mcwig.Color("statusline")
	w, h := view.Size()
	h -= 1

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	msg := fmt.Sprintf("search: %s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(0, h, msg, st)
}

func (u *uiSearchPrompt) Mode() mcwig.Mode {
	return mcwig.MODE_NORMAL
}
