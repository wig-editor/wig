package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

func HandleInsertKey(ctx Context, ev *tcell.EventKey) {
	if ctx.Buf.Mode() != MODE_INSERT {
		return
	}

	ch := ev.Rune()

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
		CmdDeleteCharBackward(ctx)
		return
	}

	line := CursorLine(ctx.Buf)

	if ev.Key() == tcell.KeyEnter {
		newLine(ctx.Buf, line)
		CmdCursorLineDown(ctx)
		CmdCursorBeginningOfTheLine(ctx)
		indent(ctx)
		return
	}

	if ctx.Buf.Cursor.Char >= len(line.Value) {
		line.Value = append(line.Value, ch)
	} else {
		tmp := []rune{ch}
		tmp = append(tmp, line.Value[ctx.Buf.Cursor.Char:]...)
		line.Value = append(line.Value[:ctx.Buf.Cursor.Char], tmp...)
		tmp = nil
	}

	if ctx.Buf.Cursor.Char < len(line.Value) {
		ctx.Buf.Cursor.Char++
		ctx.Buf.Cursor.PreserveCharPosition = ctx.Buf.Cursor.Char
	}
}
