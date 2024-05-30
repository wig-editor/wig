package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

func HandleInsertKey(e *Editor, ev *tcell.EventKey) {
	do(e, func(buf *Buffer, line *Element[Line]) {
		if buf.Mode != MODE_INSERT {
			return
		}

		ch := ev.Rune()

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

		if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
			CmdDeleteCharBackward(e)
			return
		}
		if ev.Key() == tcell.KeyEnter {
			CmdNewLine(e)
			return
		}

		if len(line.Value) == 0 {
			line.Value = append(line.Value, ch)
			CmdCursorRight(e)
			return
		}

		if buf.Cursor.Char >= len(line.Value) {
			// insert at the end of the line
			line.Value = append(line.Value, ch)
			buf.Cursor.Char++
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		} else {
			tmp := []rune{ch}
			tmp = append(tmp, line.Value[buf.Cursor.Char:]...)
			line.Value = append(line.Value[:buf.Cursor.Char], tmp...)
			CmdCursorRight(e)
		}
	})
}
