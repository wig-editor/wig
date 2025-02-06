package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

func HandleInsertKey(ctx Context, ev *tcell.EventKey) {
	if ctx.Buf.Mode() != MODE_INSERT {
		return
	}
	{
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

	ch := ev.Rune()

	if ev.Key() == tcell.KeyEnter {
		ch = '\n'
	}

	line := CursorLine(ctx.Buf)

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		start := ctx.Buf.Cursor
		start.Char--

		if start.Char < 0 {
			if line.Prev() == nil {
				return
			}

			ctx.Buf.Cursor.Line--
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
			End:   ctx.Buf.Cursor,
		})
		if ctx.Buf.Cursor.Char > 0 {
			ctx.Buf.Cursor.Char--
		}
		return
	}

	TextInsert(ctx.Buf, line, ctx.Buf.Cursor.Char, string(ch))

	if ev.Key() == tcell.KeyEnter {
		CmdCursorLineDown(ctx)
		CmdCursorBeginningOfTheLine(ctx)
		return
	}

	if ctx.Buf.Cursor.Char < len(line.Value) {
		ctx.Buf.Cursor.Char++
		ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
	}
}
