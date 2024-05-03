package mcwig

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	_ "github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
)

var msg string

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

	current := e.activeBuffer.Lines.Head
	i := 0
	y := 0
	offset := 50
	for current != nil {
		if i >= offset {
			setContent(e.screen, 0, y, fmt.Sprintf(" %d ", i+1), color("text"))
			setContent(e.screen, 4, y, string(current.Data), color("text"))
			y++
		}

		current = current.Next
		i++
	}

	e.screen.Show()
}
