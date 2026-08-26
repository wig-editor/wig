package ui

import (
	"math"
	"strings"
	"unicode"

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

	// Documentation popup state
	docText   string
	docScroll int
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
			"Up": func(ctx wig.Context) {
				widget.moveActive(-1)
			},
			"Down": func(ctx wig.Context) {
				widget.moveActive(1)
			},
			"Home": func(ctx wig.Context) {
				widget.setActive(0)
			},
			"End": func(ctx wig.Context) {
				if len(widget.items.Items) > 0 {
					widget.setActive(len(widget.items.Items) - 1)
				}
			},
			"PgUp": func(ctx wig.Context) {
				widget.moveActive(-5)
			},
			"PgDn": func(ctx wig.Context) {
				widget.moveActive(5)
			},
			"Tab": func(ctx wig.Context) {
				widget.moveActive(1)
			},
			"Backtab": func(ctx wig.Context) {
				widget.moveActive(-1)
			},
			"ctrl+d": func(ctx wig.Context) {
				widget.docScroll += 5
			},
			"ctrl+u": func(ctx wig.Context) {
				widget.docScroll -= 5
				if widget.docScroll < 0 {
					widget.docScroll = 0
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
				widget.docText = ""
				widget.docScroll = 0
				widget.items = ctx.Editor.Lsp.Completion(e.Buf)
				if len(widget.items.Items) == 0 {
					widget.Close()
				} else {
					widget.triggerDocResolve()
				}
				ctx.Editor.Redraw()
			}
		}
	}()

	ctx.Editor.PushUi(widget)
	widget.triggerDocResolve()

	return widget
}

// setActive updates the active item index and triggers documentation
// resolution for the newly selected candidate.
func (w *AutocompleteWidget) setActive(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(w.items.Items) {
		idx = len(w.items.Items) - 1
	}
	if w.activeItem == idx {
		return
	}
	w.activeItem = idx
	w.docScroll = 0
	w.triggerDocResolve()
}

// moveActive moves the active item by delta positions.
func (w *AutocompleteWidget) moveActive(delta int) {
	w.setActive(w.activeItem + delta)
}

// triggerDocResolve displays any immediately available documentation for the
// active item, then asynchronously calls completionItem/resolve to fetch
// fuller documentation from the language server.
func (w *AutocompleteWidget) triggerDocResolve() {
	if w.activeItem < 0 || w.activeItem >= len(w.items.Items) {
		w.docText = ""
		return
	}

	item := w.items.Items[w.activeItem]

	// Show immediately available docs
	if item.Documentation.Value != "" {
		w.docText = item.Documentation.Value
	} else if item.Detail != "" {
		w.docText = item.Detail
	} else {
		w.docText = ""
	}

	// Resolve fuller documentation in background
	activeIdx := w.activeItem
	buf := w.ctx.Buf
	go func() {
		doc, err := w.ctx.Editor.Lsp.CompletionItemResolve(buf, item)
		if err != nil || doc == "" {
			return
		}
		// Only update if still on the same item
		if w.activeItem != activeIdx {
			return
		}
		if doc != w.docText {
			w.docText = doc
			w.ctx.Editor.Redraw()
		}
	}()
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

	cur := wig.ContextCursorGet(ctx)
	line := wig.CursorLine(ctx.Buf, cur)
	item := w.items.Items[w.activeItem]

	if item.TextEdit == nil || (item.TextEdit.Insert.Start.Line == 0 && item.TextEdit.Insert.Start.Character == 0) {
		label := item.Label
		if label == "" {
			label = item.InsertText
		}
		cur2 := cur
		cur2.Char -= 1
		r := wig.CursorChar(ctx.Buf, cur2)
		if !unicode.IsPunct(r) && !unicode.IsSpace(r) {
			cur := wig.ContextCursorGet(ctx)
			wig.SelectionStart(ctx.Buf, cur)
			wig.WithSelection(wig.CmdBackwardWord)(ctx)
			wig.SelectionDelete(ctx)
			wig.CmdCursorLeft(ctx)
		}
		cur = wig.ContextCursorGet(ctx)
		wig.TextInsert(ctx.Buf, line, cur.Char+1, label)
		cur.Char += len(label)
		wig.CmdEnterInsertMode(ctx)
		wig.CmdCursorRight(ctx)
		return
	}

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
		cur.Char = pos
		ctx.Editor.Snippets.Expand(ctx, wig.Snippet{Body: text}, pos)
		return
	}

	chpos := len(text)
	wig.TextInsert(ctx.Buf, line, int(pos), text)
	cur.Char = item.TextEdit.Replace.Start.Character + chpos
}

func (w *AutocompleteWidget) Render(view wig.View) {
	cur := wig.ContextCursorGet(w.ctx)
	x := w.pos.Char + 2
	y := w.pos.Line - cur.ScrollOffset + 1

	maxItems := min(10, len(w.items.Items))

	vw, vh := view.Size()
	if y+maxItems >= vh {
		y -= maxItems + 2
	}

	listWidth := 50
	listX := w.pos.Char
	listY := y
	drawBoxNoBorder(view, listX, listY, listWidth, maxItems, wig.Color("ui.menu"))

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
		view.SetContent(x, listY+i, label, st)
	}

	// Render documentation popup alongside the candidate list
	w.renderDoc(view, listX, listY, maxItems, listWidth, vw)
}

// renderDoc draws a documentation popup next to the candidate list. It
// positions the popup to the right of the list when there is space,
// otherwise to the left. The popup height matches the candidate list
// height and supports vertical scrolling via docScroll.
func (w *AutocompleteWidget) renderDoc(view wig.View, listX, listY, listHeight, listWidth, vw int) {
	if w.docText == "" {
		return
	}

	docWidth := 60
	if docWidth > vw-listWidth-2 {
		docWidth = vw - listWidth - 2
	}
	if docWidth < 20 {
		return
	}

	docHeight := listHeight

	// Position: right of the list if there's space, otherwise left
	docX := listX + listWidth + 1
	if docX+docWidth > vw {
		docX = listX - docWidth - 1
	}
	if docX < 0 {
		return
	}

	wrapped := wrapText(w.docText, docWidth-2)

	if len(wrapped) < docHeight {
		docHeight = len(wrapped)
	}
	if docHeight < 1 {
		return
	}

	maxScroll := len(wrapped) - docHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if w.docScroll > maxScroll {
		w.docScroll = maxScroll
	}

	drawBoxNoBorder(view, docX, listY, docWidth, docHeight, wig.Color("ui.menu"))

	for i := 0; i < docHeight; i++ {
		idx := i + w.docScroll
		if idx >= len(wrapped) {
			break
		}
		view.SetContent(docX+1, listY+i, wrapped[idx], wig.Color("ui.menu"))
	}
}

// wrapText splits text into lines no longer than width characters,
// preferring to break at word boundaries. Each input line is wrapped
// independently so paragraph structure is preserved.
func wrapText(text string, width int) []string {
	if width < 1 {
		return []string{text}
	}
	var result []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			result = append(result, "")
			continue
		}
		runes := []rune(line)
		for len(runes) > width {
			breakAt := width
			for i := width - 1; i > 0; i-- {
				if runes[i] == ' ' {
					breakAt = i
					break
				}
			}
			result = append(result, string(runes[:breakAt]))
			if breakAt < len(runes) && runes[breakAt] == ' ' {
				runes = runes[breakAt+1:]
			} else {
				runes = runes[breakAt:]
			}
		}
		if len(runes) > 0 {
			result = append(result, string(runes))
		}
	}
	return result
}
