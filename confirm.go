package wig

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type ConfirmWidget struct {
	editor   *Editor
	keymap   *KeyHandler
	prompt   string
	onYes    func()
	onNo     func()
	onCancel func()
}

func (u *ConfirmWidget) Plane() RenderPlane  { return PlaneEditor }
func (u *ConfirmWidget) Mode() Mode          { return MODE_NORMAL }
func (u *ConfirmWidget) Keymap() *KeyHandler { return u.keymap }

func ConfirmInit(ctx Context, prompt string, onYes func(), onNo func(), onCancel func()) *ConfirmWidget {
	widget := &ConfirmWidget{
		editor:   ctx.Editor,
		prompt:   prompt,
		onYes:    onYes,
		onNo:     onNo,
		onCancel: onCancel,
	}

	km := KeyMap{
		"y": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onYes != nil {
				widget.onYes()
			}
		},
		"Y": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onYes != nil {
				widget.onYes()
			}
		},
		"Enter": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onYes != nil {
				widget.onYes()
			}
		},
		"n": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onNo != nil {
				widget.onNo()
			}
		},
		"N": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onNo != nil {
				widget.onNo()
			}
		},
		"c": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onCancel != nil {
				widget.onCancel()
			}
		},
		"C": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onCancel != nil {
				widget.onCancel()
			}
		},
		"Esc": func(ctx Context) {
			ctx.Editor.PopUi()
			if widget.onCancel != nil {
				widget.onCancel()
			}
		},
	}

	widget.keymap = NewKeyHandler(ModeKeyMap{MODE_NORMAL: km})
	ctx.Editor.PushUi(widget)
	return widget
}

func (u *ConfirmWidget) Render(view View) {
	vw, vh := view.Size()
	y := vh - 1

	// Use the statusline style to blend in with the bottom bar
	st := Color("ui.statusline")

	// Fill the entire bottom line with the background color
	bg := strings.Repeat(" ", vw)
	view.SetContent(0, y, bg, st)

	// Render the prompt text
	view.SetContent(0, y, u.prompt, st)

	// Render a red cursor block at the end of the prompt
	cursorStyle := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite)
	if len(u.prompt) < vw {
		view.SetContent(len(u.prompt), y, " ", cursorStyle)
	}
}
