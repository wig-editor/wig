package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

// TODO: optimize:
// - remove cmd calls
// - do not search for line "every" time.
func HandleInsertKey(e *Editor, ev *tcell.EventKey) {
	buf := e.ActiveWindow().Buffer()

	if buf.Mode() != MODE_INSERT {
		return
	}

	ch := ev.Rune()

	// TODO: this is why we need to send partial changes!
	defer e.Lsp.DidChange(buf)

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

	line := CursorLine(buf)

	if buf.Cursor.Char >= len(line.Value) {
		line.Value = append(line.Value, ch)
		if buf.Cursor.Char < len(line.Value) {
			buf.Cursor.Char++
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	} else {
		tmp := []rune{ch}
		tmp = append(tmp, line.Value[buf.Cursor.Char:]...)
		line.Value = append(line.Value[:buf.Cursor.Char], tmp...)
		tmp = nil
		if buf.Cursor.Char < len(line.Value) {
			buf.Cursor.Char++
			buf.Cursor.PreserveCharPosition = buf.Cursor.Char
		}
	}
}
