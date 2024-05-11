package mcwig

import "github.com/gdamore/tcell/v2"

func HandleInsertKey(e *Editor, ev *tcell.EventKey) {
	buf := e.activeBuffer
	ch := ev.Rune()

	if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
		return
	}
	if ev.Key() == tcell.KeyEnter {
		return
	}

	line := buf.Lines.Head
	for i := 0; i < buf.Cursor.Line; i++ {
		line = line.Next
	}
	line.Data = append(line.Data, ' ')
	for i := len(line.Data) - 1; i > buf.Cursor.Char; i-- {
		line.Data[i] = line.Data[i-1]
	}
	line.Data[buf.Cursor.Char] = ch

	CmdCursorRight(e)
}
