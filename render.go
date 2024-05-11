package mcwig

import (
	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

func setContent(s tcell.Screen, x, y int, str string, st tcell.Style) int {
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
	e.screen.Clear()
	// e.screen.Fill(0, color("bg"))

	currentLine := e.activeBuffer.Lines.Head
	lineNum := 0
	y := 0
	offset := e.activeBuffer.ScrollOffset
	for currentLine != nil {
		if lineNum >= offset {
			// setContent(e.screen, 0, y, fmt.Sprintf(" %d|", lineNum), color("text"))

			if lineNum == e.activeBuffer.Cursor.Line {
				// render text
				setContent(e.screen, 0, y, string(currentLine.Data), color("text"))
				if len(currentLine.Data) > 0 {
					// render cursor
					ch := '•'
					curPos := e.activeBuffer.Cursor.Char
					if curPos < len(currentLine.Data) {
						ch = currentLine.Data[curPos]
					}
					setContent(e.screen, e.activeBuffer.Cursor.Char, y, string(ch), color("cursor"))
				} else {
					// render cursor on empty line
					setContent(e.screen, 0, y, " ", color("cursor"))
				}
			} else {
				// render text
				setContent(e.screen, 0, y, string(currentLine.Data), color("text"))
			}

			y++
		}

		currentLine = currentLine.Next
		lineNum++
	}

	e.screen.Show()
}
