package wig

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

func HandleInsertKey(ctx Context, ev *tcell.EventKey) {
	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	ch := ev.Rune()

	if ev.Key() == tcell.KeyCtrlJ {
		ev = tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	}

	{
		if ctx.Buf.Mode() != MODE_INSERT {
			return
		}
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			return
		}
		if ev.Modifiers()&tcell.ModAlt != 0 {
			return
		}
		if ev.Modifiers()&tcell.ModMeta != 0 {
			return
		}
	}

	if ev.Key() == tcell.KeyEnter {
		ch = '\n'
	}

	{
		if ch == '\t' {
			if Tabstopped(ctx) {
				TabstopNext(ctx)
				return
			}
			if strings.TrimSpace(line.Value.String()) == "" {
				goto insertChar
			}
			if cur.Char >= len(line.Value.String())-1 {
				if ctx.Editor.AutocompleteTrigger(ctx) {
					return
				}
			}
		}
	}

insertChar:

	if ch == 0 {
		return
	}

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		start := *cur
		start.Char--

		if start.Char < 0 {
			if line.Prev() == nil {
				return
			}

			cur.Line--
			CmdGotoLineEnd(ctx)

			// delete \n on prev line
			TextDelete(ctx.Buf, &Selection{
				Start: Cursor{Line: start.Line - 1, Char: len(line.Prev().Value) - 1},
				End:   Cursor{Line: start.Line - 1, Char: len(line.Prev().Value)},
			})

			return
		}

		TextDelete(ctx.Buf, &Selection{
			Start: start,
			End:   *cur,
		})
		if cur.Char > 0 {
			cur.Char--
		}
		return
	}

	SelectionDelete(ctx)
	TextInsert(ctx.Buf, line, cur.Char, string(ch))

	if ev.Key() == tcell.KeyEnter {
		CmdCursorLineDown(ctx)
		CmdCursorBeginningOfTheLine(ctx)
		return
	}

	if cur.Char < len(line.Value) {
		cur.Char++
		cur.PreserveCharPosition = cur.Char
	}
}

