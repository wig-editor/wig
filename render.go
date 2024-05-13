package mcwig

import (
	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

func SetContent(s tcell.Screen, x, y int, str string, st tcell.Style) int {
	xx := x
	for _, ch := range str {
		var comb []rune
		w := runewidth.RuneWidth(ch)
		if w == 0 {
			comb = []rune{ch}
			ch = ' '
			w = 1
		}

		s.SetContent(xx, y, ch, comb, st)
		xx += w
	}
	return xx - x
}

func (e *Editor) render() {
	e.Screen.Clear()

	buf := e.ActiveBuffer
	currentLine := buf.Lines.Head
	lineNum := 0
	y := 0
	offset := buf.ScrollOffset
	for currentLine != nil {
		if lineNum >= offset {
			// render each character in the line separately
			x := 0

			if len(currentLine.Data) == 0 && lineNum == buf.Cursor.Line {
				SetContent(e.Screen, x, y, " ", Color("cursor"))
			}

			for i := 0; i < len(currentLine.Data); i++ {
				ch := getRenderChar(currentLine.Data[i])
				SetContent(e.Screen, x, y, string(ch), Color("text"))

				// render cursor
				if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
					SetContent(e.Screen, x, y, string(ch[0]), Color("cursor"))
				}

				x += len(ch)
			}

			y++
		}

		currentLine = currentLine.Next
		lineNum++
	}

	// render widgets
	for _, w := range e.widgets {
		w.Render()
	}

	e.Screen.Show()
}

func getRenderChar(ch rune) string {
	if ch == '\t' {
		return "    "
	}
	return string(ch)
}
