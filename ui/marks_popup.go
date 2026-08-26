package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

type MarksPopupWidget struct {
	e      *wig.Editor
	keymap *wig.KeyHandler
	items  []markItem
}

type markItem struct {
	Mark rune
	Line int
	Text string
}

func (u *MarksPopupWidget) Plane() wig.RenderPlane  { return wig.PlaneEditor }
func (u *MarksPopupWidget) Mode() wig.Mode          { return wig.MODE_NORMAL }
func (u *MarksPopupWidget) Keymap() *wig.KeyHandler { return u.keymap }

func MarksPopupInit(ctx wig.Context, marks map[rune]wig.Cursor) {
	widget := &MarksPopupWidget{
		e: ctx.Editor,
	}

	var keys []rune
	for k := range marks {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	km := wig.KeyMap{
		"Esc": func(ctx wig.Context) {
			ctx.Editor.PopUi()
		},
		// Double backtick / quote (pingpong): jump back and forward
		"`": func(ctx wig.Context) {
			ctx.Editor.PopUi()
			wig.CmdJumpToggle(ctx)
		},
		"'": func(ctx wig.Context) {
			ctx.Editor.PopUi()
			wig.CmdJumpToggle(ctx)
		},
	}

	for _, k := range keys {
		cur := marks[k]
		line := wig.CursorLineByNum(ctx.Buf, cur.Line)
		text := ""
		if line != nil {
			text = strings.TrimRight(line.Value.String(), "\n")
			text = strings.TrimSpace(text)
			if len(text) > 60 {
				text = text[:60] + "..."
			}
		}

		mark := k
		targetCur := cur
		widget.items = append(widget.items, markItem{
			Mark: mark,
			Line: cur.Line + 1,
			Text: text,
		})

		jumpFn := func(ctx wig.Context) {
			ctx.Editor.PopUi()
			win := ctx.Win
			if win == nil {
				win = ctx.Editor.ActiveWindow()
			}
			win.VisitBuffer(ctx, targetCur)
			wig.CmdEnsureCursorVisible(ctx)
		}

		markStr := string(mark)
		km[markStr] = jumpFn
		km["shift+"+markStr] = jumpFn
	}

	widget.keymap = wig.NewKeyHandler(wig.ModeKeyMap{wig.MODE_NORMAL: km})
	ctx.Editor.PushUi(widget)
}

func (u *MarksPopupWidget) Render(view wig.View) {
	vw, vh := view.Size()

	lines := []string{}
	if len(u.items) == 0 {
		lines = append(lines, " No marks set. Press ` for pingpong, Esc to close. ")
	} else {
		for _, item := range u.items {
			s := fmt.Sprintf(" '%c'  line %-4d │ %s ", item.Mark, item.Line, item.Text)
			lines = append(lines, s)
		}
	}

	boxH := len(lines) + 1
	boxW := int(float32(vw) * 0.85)
	if boxW > vw-4 {
		boxW = vw - 4
	}
	if boxW < 40 {
		boxW = min(vw, 40)
	}

	x := (vw - boxW) / 2
	if x < 0 {
		x = 0
	}
	y := vh - boxH - 2
	if y < 0 {
		y = 0
	}

	style := wig.Color("default")
	drawBox(view, x, y, x+boxW, y+boxH, style)

	titleStyle := wig.Color("ui.popup.title")
	if titleStyle == style {
		titleStyle = wig.Color("ui.linenr.selected")
	}
	view.SetContent(x+2, y, " Marks (` or ' for pingpong) ", titleStyle)

	textStyle := wig.Color("default")
	markStyle := wig.Color("ui.mark")
	if markStyle == style {
		markStyle = style.Foreground(tcell.ColorYellow).Bold(true)
	}
	lineNrStyle := wig.Color("ui.linenr")

	if len(u.items) == 0 {
		for i, line := range lines {
			view.SetContent(x+2, y+1+i, line, textStyle)
		}
	} else {
		for i, item := range u.items {
			cx := x + 2
			row := y + 1 + i

			view.SetContent(cx, row, "'", textStyle)
			cx++

			view.SetContent(cx, row, string(item.Mark), markStyle)
			cx += len([]rune(string(item.Mark)))

			view.SetContent(cx, row, "'  line ", textStyle)
			cx += 8

			lnStr := fmt.Sprintf("%-4d", item.Line)
			view.SetContent(cx, row, lnStr, lineNrStyle)
			cx += len(lnStr)

			view.SetContent(cx, row, " │ ", textStyle)
			cx += 3

			avail := (x + boxW - 1) - cx
			if avail > 0 {
				view.SetContent(cx, row, truncate(item.Text, avail), textStyle)
			}
		}
	}
}
