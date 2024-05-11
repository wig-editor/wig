package mcwig

import (
	"github.com/gdamore/tcell/v2"
)

func HandleInsertKey(e *Editor, ev *tcell.EventKey) {
	buf := e.activeBuffer
	ch := ev.Rune()

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		CmdDeleteCharBackward(e)
		return
	}
	if ev.Key() == tcell.KeyEnter {
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

	tmp := []rune{ch}
	tmp = append(tmp, line.Data[buf.Cursor.Char:]...)
	line.Data = append(line.Data[:buf.Cursor.Char], tmp...)

	CmdCursorRight(e)
}
