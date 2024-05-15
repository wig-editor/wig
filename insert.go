package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

func HandleInsertKey(e *Editor, ev *tcell.EventKey) {
	buf := e.ActiveBuffer
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

	line := buf.Lines.Head
	for i := 0; i < buf.Cursor.Line; i++ {
		line = line.Next
	}

	if len(line.Data) == 0 {
		line.Data = append(line.Data, ch)
		CmdCursorRight(e)
		return
	}

	if buf.Cursor.Char >= len(line.Data) {
		// insert at the end of the line
		line.Data = append(line.Data, ch)
		buf.Cursor.Char++
		buf.Cursor.PreserveCharPosition = buf.Cursor.Char
	} else {
		tmp := []rune{ch}
		tmp = append(tmp, line.Data[buf.Cursor.Char:]...)
		line.Data = append(line.Data[:buf.Cursor.Char], tmp...)
		CmdCursorRight(e)
	}
}
