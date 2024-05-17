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
	offset := buf.ScrollOffset
	lineNum := 0
	y := 0
	for currentLine != nil {
		if lineNum >= offset {
			// render each character in the line separately
			x := 0

			// render cursor on empty line
			if len(currentLine.Data) == 0 && lineNum == buf.Cursor.Line {
				SetContent(e.Screen, x, y, " ", Color("cursor"))
			}

			for i := 0; i < len(currentLine.Data); i++ {

				// render selection
				textStyle := Color("text")
				if buf.Selection != nil {
					if SelectionCursorInRange(buf.Selection, Cursor{Line: lineNum, Char: i}) {
						textStyle = Color("statusline.normal")
					}
				}

				ch := getRenderChar(currentLine.Data[i])
				SetContent(e.Screen, x, y, string(ch), textStyle)

				// render cursor
				if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
					SetContent(e.Screen, x, y, string(ch[0]), Color("cursor"))
				}

				x += len(ch)
			}

			// render cursor after the end of the line in insert mode
			if lineNum == buf.Cursor.Line && buf.Cursor.Char >= len(currentLine.Data) {
				SetContent(e.Screen, x, y, " ", Color("cursor"))
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
