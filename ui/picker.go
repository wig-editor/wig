package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/firstrow/mcwig"
)

type PickerItem[T any] struct {
	Name  string
	Value T
}

type PickerAction[T any] func(i PickerItem[T])

type uiPicker[T any] struct {
	e          *mcwig.Editor
	keymap     *mcwig.KeyHandler
	items      []PickerItem[T]
	action     PickerAction[T]
	chBuf      []rune
	activeItem int
}

func PickerInit[T any](
	e *mcwig.Editor,
	action PickerAction[T],
	items []PickerItem[T],
) {
	picker := &uiPicker[T]{
		e:      e,
		chBuf:  []rune{},
		items:  items,
		action: action,
	}
	picker.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			"Esc": func(e *mcwig.Editor) {
				e.PopUi()
			},
			"Tab": func(e *mcwig.Editor) {
				e.PopUi()
			},
			"Enter": func(e *mcwig.Editor) {

			},
		},
	})
	picker.keymap.Fallback(picker.insertCh)

	e.PushUi(picker)
}

func (u *uiPicker[T]) insertCh(e *mcwig.Editor, ev *tcell.EventKey) {
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
		}
		return
	}
	if ev.Key() == tcell.KeyEnter {
		// text := strings.TrimSpace(string(u.chBuf))

		e.PopUi()
		return
	}

	u.chBuf = append(u.chBuf, ev.Rune())
}

func (u *uiPicker[T]) Mode() mcwig.Mode {
	return mcwig.MODE_NORMAL
}

func (u *uiPicker[T]) Keymap() *mcwig.KeyHandler {
	return u.keymap
}

func (u *uiPicker[T]) Render(view mcwig.View, viewport mcwig.Viewport) {
	vw, vh := viewport.Size()

	w := int(float32(vw) * 0.8)
	h := vh - 5
	x := vw/2 - w/2
	y := 3

	drawBox(view, x, y, w, h, mcwig.Color("default"))

	// prompt
	prompt := fmt.Sprintf(" %s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(x+1, y+1, prompt, mcwig.Color("default"))

	// separator
	line := strings.Repeat(string(tcell.RuneHLine), w-x-3)
	view.SetContent(x+2, y+2, line, mcwig.Color("default"))

	i := 3

	// TODO: handle 100s of items
	for _, row := range u.items {
		line = fmt.Sprintf("> %s", truncate(row.Name, w-x-5))
		view.SetContent(x+2, y+i, line, mcwig.Color("default"))
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-3]) + "..."
}

func drawBox(s mcwig.View, x1, y1, x2, y2 int, style tcell.Style) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, string(tcell.RuneHLine), style)
		s.SetContent(col, y2, string(tcell.RuneHLine), style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, string(tcell.RuneVLine), style)
		s.SetContent(x2, row, string(tcell.RuneVLine), style)
	}
	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		s.SetContent(x1, y1, string(tcell.RuneULCorner), style)
		s.SetContent(x2, y1, string(tcell.RuneURCorner), style)
		s.SetContent(x1, y2, string(tcell.RuneLLCorner), style)
		s.SetContent(x2, y2, string(tcell.RuneLRCorner), style)
	}
	// fill
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			s.SetContent(col, row, " ", style)
		}
	}
}
