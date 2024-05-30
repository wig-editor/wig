package ui

import "github.com/firstrow/mcwig"

func WindowRender(e *mcwig.Editor, view mcwig.View, win *mcwig.Window) {
	buf := win.Buffer
	if buf == nil {
		return
	}

	currentLine := buf.Lines.First()
	offset := buf.ScrollOffset
	lineNum := 0
	y := 0

	for currentLine != nil {
		if lineNum >= offset {
			// render each character in the line separately
			x := 0

			// render cursor on empty line
			if len(currentLine.Value) == 0 && lineNum == buf.Cursor.Line {
				view.SetContent(x, y, " ", mcwig.Color("cursor"))
			}

			for i := 0; i < len(currentLine.Value); i++ {
				// render selection
				textStyle := mcwig.Color("default")
				if buf.Selection != nil {
					if mcwig.SelectionCursorInRange(buf.Selection, mcwig.Cursor{Line: lineNum, Char: i}) {
						textStyle = mcwig.Color("statusline")
					}
				}

				ch := getRenderChar(currentLine.Value[i])
				view.SetContent(x, y, string(ch), textStyle)

				// render cursor
				if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
					view.SetContent(x, y, string(ch[0]), mcwig.Color("cursor"))
				}

				x += len(ch)
			}

			// render cursor after the end of the line in insert mode
			if lineNum == buf.Cursor.Line && buf.Cursor.Char >= len(currentLine.Value) {
				view.SetContent(x, y, " ", mcwig.Color("cursor"))
			}

			y++
		}

		currentLine = currentLine.Next()
		lineNum++
	}
}

func getRenderChar(ch rune) string {
	if ch == '\t' {
		return "    "
	}
	return string(ch)
}
