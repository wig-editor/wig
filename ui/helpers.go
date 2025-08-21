package ui

import (
	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-3]) + "..."
}

func drawBox(s wig.View, x1, y1, x2, y2 int, style tcell.Style) {
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, string(tcell.RuneHLine), style)
		s.SetContent(col, y2, string(tcell.RuneHLine), style)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, string(tcell.RuneVLine), style)
		s.SetContent(x2, row, string(tcell.RuneVLine), style)
	}
	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		s.SetContent(x1, y1, string(tcell.RuneULCorner), style)
		s.SetContent(x2, y1, string(tcell.RuneURCorner), style)
		s.SetContent(x1, y2, string(tcell.RuneLLCorner), style)
		s.SetContent(x2, y2, string(tcell.RuneLRCorner), style)
	}

	// fill bg
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			s.SetContent(col, row, " ", style)
		}
	}
}

func drawBox2(s wig.View, x, y, width, height int, style tcell.Style) {
	drawBox(s, x, y, x+width, y+height, style)
}

func drawBoxNoBorder(s wig.View, x1, y1, width, height int, style tcell.Style) {
	x2 := x1 + width
	y2 := y1 + height
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for row := y1; row < y2; row++ {
		for col := x1; col < x2; col++ {
			s.SetContent(col, row, " ", style)
		}
	}
}

