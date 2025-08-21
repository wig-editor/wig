package ui

import (
	"fmt"
	"strings"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

type uiSearchPrompt struct {
	e      *wig.Editor
	keymap *wig.KeyHandler
	chBuf  []rune
}

func (u *uiSearchPrompt) Plane() wig.RenderPlane {
	return wig.PlaneEditor
}

func CmdSearchPromptInit(ctx wig.Context) {
	cmdLine := &uiSearchPrompt{
		e:     ctx.Editor,
		chBuf: []rune{},
	}

	cmdLine.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Esc": func(ctx wig.Context) {
				ctx.Editor.PopUi()
			},
		},
	})
	cmdLine.keymap.Fallback(cmdLine.insertCh)
	ctx.Editor.PushUi(cmdLine)
}

func (u *uiSearchPrompt) insertCh(ctx wig.Context, ev *tcell.EventKey) {
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
	wig.LastSearchPattern = pat
	wig.SearchNext(u.e.NewContext(), pat)
}

func (u *uiSearchPrompt) Keymap() *wig.KeyHandler {
	return u.keymap
}

func (u *uiSearchPrompt) Render(view wig.View) {
	st := wig.Color("statusline")
	w, h := view.Size()
	h -= 1

	bg := strings.Repeat(" ", w)
	view.SetContent(0, h, bg, st)

	msg := fmt.Sprintf("search: %s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(0, h, msg, st)
}

func (u *uiSearchPrompt) Mode() wig.Mode {
	return wig.MODE_NORMAL
}

