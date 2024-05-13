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

	currentLine := e.activeBuffer.Lines.Head
	lineNum := 0
	y := 0
	offset := e.activeBuffer.ScrollOffset
	for currentLine != nil {
		if lineNum >= offset {
			// render each character in the line separately
			x := 0

			if len(currentLine.Data) == 0 && lineNum == e.activeBuffer.Cursor.Line {
				setContent(e.screen, x, y, " ", color("cursor"))
			}

			for i := 0; i < len(currentLine.Data); i++ {
				ch := getRenderChar(currentLine.Data[i])
				setContent(e.screen, x, y, string(ch), color("text"))

				// render cursor
				if lineNum == e.activeBuffer.Cursor.Line && i == e.activeBuffer.Cursor.Char {
					setContent(e.screen, x, y, string(ch[0]), color("cursor"))
				}

				x += len(ch)
			}

			y++
		}

		currentLine = currentLine.Next
		lineNum++
	}

	e.screen.Show()
}

func getRenderChar(ch rune) string {
	if ch == '\t' {
		return "    "
	}
	return string(ch)
}
