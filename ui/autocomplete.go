package ui

import (
	"math"

	"github.com/firstrow/wig"
)

type AutocompleteWidget struct {
	ctx            wig.Context
	triggerPos     wig.Cursor
	keymap         *wig.KeyHandler
	pos            wig.Position
	items          wig.CompletionItems
	eventsListener <-chan wig.Event
	activeItem     int
}

func (u *AutocompleteWidget) Plane() wig.RenderPlane {
	return wig.PlaneWin
}

func AutocompleteInit(
	ctx wig.Context,
	pos wig.Position,
	items wig.CompletionItems,
) *AutocompleteWidget {
	if len(items.Items) == 0 {
		return nil
	}

	widget := &AutocompleteWidget{
		ctx:        ctx,
		pos:        pos,
		items:      items,
		activeItem: 0,
	}

	widget.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_INSERT: wig.KeyMap{
			"Esc": func(ctx wig.Context) {
				widget.Close()
			},
			"Tab": func(ctx wig.Context) {
				if widget.activeItem < len(widget.items.Items)-1 {
					widget.activeItem++
				}
			},
			"Backtab": func(ctx wig.Context) {
				if widget.activeItem > 0 {
					widget.activeItem--
				}
			},
			"Enter": widget.selectItem,
		},
	})

	widget.eventsListener = ctx.Editor.Events.Subscribe()
	go func() {
		for event := range widget.eventsListener {
			event.Wg.Done()
			switch e := event.Msg.(type) {
			case wig.EventTextChange:
				widget.activeItem = 0
				widget.items = ctx.Editor.Lsp.Completion(e.Buf)
				if len(widget.items.Items) == 0 {
					widget.Close()
				}
				ctx.Editor.Redraw()
			}
		}
	}()

	ctx.Editor.PushUi(widget)

	return widget
}

func (w *AutocompleteWidget) Close() {
	w.ctx.Editor.PopUi()
	w.ctx.Editor.Events.Unsubscribe(w.eventsListener)
	w.ctx.Editor.Redraw()
}

func (w *AutocompleteWidget) Mode() wig.Mode {
	return wig.MODE_INSERT
}

func (w *AutocompleteWidget) Keymap() *wig.KeyHandler {
	return w.keymap
}

func (w *AutocompleteWidget) selectItem(ctx wig.Context) {
	defer w.Close()

	line := wig.CursorLine(ctx.Buf)
	item := w.items.Items[w.activeItem]
	text := item.TextEdit.NewText
	pos := item.TextEdit.Insert.Start.Character

	wig.TextDelete(ctx.Buf, &wig.Selection{
		Start: wig.Cursor{
			Line: item.TextEdit.Replace.Start.Line,
			Char: item.TextEdit.Replace.Start.Character,
		},
		End: wig.Cursor{
			Line: item.TextEdit.Replace.End.Line,
			Char: item.TextEdit.Replace.End.Character,
		},
	})

	if item.InsertTextFormat == 2 {
		ctx.Buf.Cursor.Char = pos
		ctx.Editor.Snippets.Expand(ctx, wig.Snippet{Body: text})
		return
	}

	chpos := len(text)
	wig.TextInsert(ctx.Buf, line, int(pos), text)
	ctx.Buf.Cursor.Char = item.TextEdit.Replace.Start.Character + chpos
}

func (w *AutocompleteWidget) Render(view wig.View) {
	x := w.pos.Char + 2
	y := w.pos.Line - w.ctx.Buf.ScrollOffset + 1

	maxItems := min(10, len(w.items.Items))

	_, winHeight := view.Size()
	if y+maxItems >= winHeight {
		y -= maxItems + 2
	}

	drawBoxNoBorder(view, w.pos.Char, y, 50, maxItems, wig.Color("ui.menu"))

	// pagination
	pageSize := maxItems
	pageNumber := math.Ceil(float64(w.activeItem+1)/float64(pageSize)) - 1
	startIndex := int(pageNumber) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > len(w.items.Items) {
		endIndex = len(w.items.Items)
	}
	dataset := w.items.Items[startIndex:endIndex]

	for i, row := range dataset {
		st := wig.Color("ui.menu")
		if i+startIndex == w.activeItem {
			st = wig.Color("ui.menu.selected")
		}

		label := row.Label
		view.SetContent(x, y, label, st)
		if i >= maxItems {
			return
		}
		y++
	}
}

