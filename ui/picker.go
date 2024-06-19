package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/sahilm/fuzzy"

	"github.com/firstrow/mcwig"
)

type PickerItem[T any] struct {
	Name   string
	Value  T
	Active bool
	Picker *UiPicker[T]
}

type PickerAction[T any] func(p *UiPicker[T], i *PickerItem[T])

type UiPicker[T any] struct {
	e           *mcwig.Editor
	keymap      *mcwig.KeyHandler
	items       []PickerItem[T]
	filtered    []PickerItem[T]
	action      PickerAction[T]
	chBuf       []rune
	activeItem  int
	activeItemT *PickerItem[T]
}

func PickerInit[T any](e *mcwig.Editor, action PickerAction[T], items []PickerItem[T]) *UiPicker[T] {
	picker := &UiPicker[T]{
		e:          e,
		chBuf:      []rune{},
		items:      items,
		filtered:   items,
		action:     action,
		activeItem: 0,
	}
	picker.keymap = mcwig.NewKeyHandler(mcwig.ModeKeyMap{
		mcwig.MODE_NORMAL: mcwig.KeyMap{
			"Esc": func(e *mcwig.Editor) {
				e.PopUi()
			},
			"Tab": func(e *mcwig.Editor) {
				if picker.activeItem < len(picker.filtered)-1 {
					picker.activeItem++
				}
			},
			"Backtab": func(e *mcwig.Editor) {
				if picker.activeItem > 0 {
					picker.activeItem--
				}
			},
			"Enter": func(e *mcwig.Editor) {
				picker.action(picker, picker.activeItemT)
			},
		},
	})
	picker.keymap.Fallback(picker.insertCh)
	e.PushUi(picker)
	return picker
}

func (u *UiPicker[T]) Mode() mcwig.Mode {
	return mcwig.MODE_NORMAL
}

func (u *UiPicker[T]) Keymap() *mcwig.KeyHandler {
	return u.keymap
}

func (u *UiPicker[T]) SetItems(items []PickerItem[T]) {
	u.items = items
	u.filtered = items
	u.chBuf = u.chBuf[:0]
	u.activeItem = 0
}

func (u *UiPicker[T]) insertCh(e *mcwig.Editor, ev *tcell.EventKey) {
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
			u.filterItems()
		}
		return
	}
	if ev.Key() == tcell.KeyEnter {
		e.PopUi()
		return
	}

	u.chBuf = append(u.chBuf, ev.Rune())
	u.filterItems()
}

func (u *UiPicker[T]) filterItems() {
	pattern := string(u.chBuf)
	if len(pattern) == 0 {
		u.filtered = u.items
		return
	}

	data := make([]string, 0, len(u.items))

	for _, row := range u.items {
		data = append(data, row.Name)
	}

	matches := fuzzy.Find(pattern, data)
	u.filtered = make([]PickerItem[T], 0, len(matches))
	for _, row := range matches {
		u.filtered = append(u.filtered, u.items[row.Index])
	}

	u.activeItem = 0
}

func (u *UiPicker[T]) Render(view mcwig.View) {
	vw, vh := view.Size()

	w := int(float32(vw) * 0.86)
	h := vh - 5
	x := vw/2 - w/2
	y := 3
	pageSize := h - 6

	// fill box
	drawBox(view, x, y, w, h, mcwig.Color("default"))

	// prompt
	prompt := fmt.Sprintf(" %s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(x+1, y+1, prompt, mcwig.Color("default"))

	// separator
	line := strings.Repeat(string(tcell.RuneHLine), w-x-3)
	view.SetContent(x+2, y+2, line, mcwig.Color("default"))

	// pagination
	pageNumber := math.Ceil(float64(u.activeItem+1)/float64(pageSize)) - 1
	startIndex := int(pageNumber) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > len(u.filtered) {
		endIndex = len(u.filtered)
	}

	dataset := u.filtered[startIndex:endIndex]

	u.activeItemT = nil

	i := 0
	for key, row := range dataset {
		isCurrent := " "
		if row.Active {
			isCurrent = "*"
		}

		if key+startIndex == u.activeItem {
			u.activeItemT = &row
			line = fmt.Sprintf("> %s %s", isCurrent, truncate(row.Name, w-x-5))
		} else {
			line = fmt.Sprintf("  %s %s", isCurrent, truncate(row.Name, w-x-5))
		}
		view.SetContent(x+2, y+i+3, line, mcwig.Color("default"))
		i++
	}
}
