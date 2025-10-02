package ui

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"

	"github.com/firstrow/wig"

	"github.com/gdamore/tcell/v2"
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

type Location struct {
	FilePath string
	Line uint
	Char uint
}

type PickerItem[T any] struct {
	Name   string
	Value  T
	Active bool
	Location Location
	Score  int
}

type PickerAction[T any] func(p *UiPicker[T], i *PickerItem[T])

type UiPicker[T any] struct {
	e           *wig.Editor
	keymap      *wig.KeyHandler
	items       []PickerItem[T]
	filtered    []PickerItem[T]
	action      PickerAction[T]
	chBuf       []rune
	activeItem  int
	activeItemT *PickerItem[T]
	onChange    func()               // on user change input
	onSelect    func(*PickerItem[T]) // when Tab pressed
}

func (u *UiPicker[T]) Plane() wig.RenderPlane {
	return wig.PlaneEditor
}

func PickerInit[T any](e *wig.Editor, action PickerAction[T], items []PickerItem[T]) *UiPicker[T] {
	for i := range items {
		name := strings.TrimRightFunc(items[i].Name, unicode.IsSpace)
		items[i].Name = strings.ReplaceAll(name, "\t", "    ")
	}

	picker := &UiPicker[T]{
		e:          e,
		chBuf:      make([]rune, 0, 32),
		items:      items,
		filtered:   items,
		action:     action,
		activeItem: 0,
		onSelect:   func(*PickerItem[T]) {},
		onChange:   func() {},
	}
	picker.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_INSERT: wig.KeyMap{
			"Esc": func(ctx wig.Context) {
				ctx.Editor.PopUi()
			},
			"Tab": func(ctx wig.Context) {
				if picker.activeItem < len(picker.filtered)-1 {
					picker.activeItem++
					picker.onSelect(&picker.filtered[picker.activeItem])
				}
			},
			"Backtab": func(ctx wig.Context) {
				if picker.activeItem > 0 {
					picker.activeItem--
					picker.onSelect(&picker.filtered[picker.activeItem])
				}
			},
			"Enter": func(ctx wig.Context) {
				picker.action(picker, picker.activeItemT)
			},
		},
	})
	picker.keymap.Fallback(picker.insertCh)
	e.PushUi(picker)
	return picker
}

func (u *UiPicker[T]) Mode() wig.Mode {
	return wig.MODE_INSERT
}

func (u *UiPicker[T]) Keymap() *wig.KeyHandler {
	return u.keymap
}

func (u *UiPicker[T]) OnChange(callback func()) {
	u.onChange = callback
}

func (u *UiPicker[T]) OnKey(key string, f func(wig.Context)) {
	newMap := wig.KeyMap{}
	newMap[key] = f
	u.keymap.Map(u.e, wig.MODE_INSERT, newMap)
}

func (u *UiPicker[T]) OnSelect(callback func(*PickerItem[T])) {
	u.onSelect = callback
}

func (u *UiPicker[T]) CallAction() {
	u.action(u, u.activeItemT)
}

func (u *UiPicker[T]) SetItems(items []PickerItem[T]) {
	for i, _ := range items {
		name := strings.TrimRightFunc(items[i].Name, unicode.IsSpace)
		items[i].Name = strings.ReplaceAll(name, "\t", "    ")
	}

	u.items = items
	u.filtered = items
	u.activeItem = 0
}

func (u *UiPicker[T]) ClearInput() {
	u.chBuf = u.chBuf[:0]
	u.activeItem = 0
}

func (u *UiPicker[T]) insertCh(ctx wig.Context, ev *tcell.EventKey) {
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
			u.onChange()
			u.filterItems()
		}
		return
	}
	if ev.Key() == tcell.KeyEnter {
		ctx.Editor.PopUi()
		return
	}

	u.chBuf = append(u.chBuf, ev.Rune())
	u.onChange()
	u.filterItems()
}

func (u *UiPicker[T]) filterItems() {
	pattern := string(u.chBuf)
	if len(pattern) == 0 {
		u.filtered = u.items
		return
	}

	u.filtered = make([]PickerItem[T], 0, len(u.items))
	pattern = strings.ReplaceAll(pattern, " ", "")
	pattern = strings.ToLower(pattern)

	for i, row := range u.items {
		chars := util.ToChars([]byte(row.Name))
		res, _ := algo.FuzzyMatchV1(false, false, true, &chars, []rune(pattern), true, nil)
		if res.Start >= 0 {
			item := u.items[i]
			item.Score = res.Score
			u.filtered = append(u.filtered, item)
		}
	}

	slices.SortFunc(u.filtered, func(a, b PickerItem[T]) int {
		return cmp.Compare(b.Score, a.Score)
	})
	u.activeItem = 0
}

func (u *UiPicker[T]) GetInput() string {
	return string(u.chBuf)
}

func (u *UiPicker[T]) SetInput(val string) {
	u.chBuf = []rune(val)
}

func (u *UiPicker[T]) Render(view wig.View) {
	vw, vh := view.Size()

	w := int(float32(vw) * 0.76)
	h := vh - 5
	x := vw/2 - w/2
	y := 3
	pageSize := h - 6

	// fill box
	drawBox(view, x, y, w, h, wig.Color("default"))

	// prompt
	prompt := fmt.Sprintf(" %s%s", string(u.chBuf), string(tcell.RuneBlock))
	view.SetContent(x+1, y+1, prompt, wig.Color("default"))

	// separator
	line := strings.Repeat(string(tcell.RuneHLine), w-x-3)
	view.SetContent(x+2, y+2, line, wig.Color("default"))

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
			line = fmt.Sprintf("> %s %s", isCurrent, truncate(row.Name, w-x-8))
		} else {
			line = fmt.Sprintf("  %s %s", isCurrent, truncate(row.Name, w-x-8))
		}
		view.SetContent(x+2, y+i+3, line, wig.Color("default"))
		i++
	}
}

