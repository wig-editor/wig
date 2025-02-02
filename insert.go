package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

// TODO: optimize:
// - remove cmd calls
// - do not search for line "every" time.
func HandleInsertKey(ctx Context, ev *tcell.EventKey) {
	if ctx.Buf.Mode() != MODE_INSERT {
		return
	}

	ch := ev.Rune()

	// TODO: this is why we need to send partial changes!
	// defer e.Lsp.DidChange(buf)

	// check for CTRL modifier
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		return
	}

	if ev.Modifiers()&tcell.ModAlt != 0 {
		return
	}

	if ev.Modifiers()&tcell.ModMeta != 0 {
		return
	}

	line := CursorLine(ctx.Buf)

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		CmdDeleteCharBackward(ctx)
		return
	}
	if ev.Key() == tcell.KeyEnter {
		newLine(ctx.Buf, line)
		CmdCursorLineDown(ctx)
		CmdCursorBeginningOfTheLine(ctx)
		indent(ctx)
		return
	}

	if ctx.Buf.Cursor.Char >= len(line.Value) {
		line.Value = append(line.Value, ch)
		if ctx.Buf.Cursor.Char < len(line.Value) {
			ctx.Buf.Cursor.Char++
			ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
		}
	} else {
		tmp := []rune{ch}
		tmp = append(tmp, line.Value[ctx.Buf.Cursor.Char:]...)
		line.Value = append(line.Value[:ctx.Buf.Cursor.Char], tmp...)
		tmp = nil
		if ctx.Buf.Cursor.Char < len(line.Value) {
			ctx.Buf.Cursor.Char++
			ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
		}
	}
}
