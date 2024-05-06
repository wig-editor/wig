package mcwig

import (
	"fmt"

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
			setContent(e.screen, 0, y, fmt.Sprintf(" %d ", lineNum), color("text"))

			if lineNum == e.activeBuffer.Cursor.Line {
				setContent(e.screen, 4, y, string(currentLine.Data), color("text"))
				if len(currentLine.Data) > 0 {
					setContent(e.screen, 4+e.activeBuffer.Cursor.Char, y, string(currentLine.Data[e.activeBuffer.Cursor.Char]), color("cursor"))
				} else {
					setContent(e.screen, 4, y, " ", color("cursor"))
				}
			} else {
				setContent(e.screen, 4, y, string(currentLine.Data), color("text"))
			}

			y++
		}

		currentLine = currentLine.Next
		lineNum++
	}

	e.screen.Show()
}
