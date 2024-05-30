package ui

import "github.com/firstrow/mcwig"

func WindowRender(e *mcwig.Editor, view mcwig.View, win *mcwig.Window) {
	buf := win.Buffer
	if buf == nil {
		return
	}

	width, _ := view.Size()
	width -= 2

	currentLine := buf.Lines.First()
	offset := buf.ScrollOffset
	lineNum := 0
	y := 0

	isActiveWin := win == e.ActiveWindow()

	skip := 0
	if buf.Cursor.Char > width {
		skip = buf.Cursor.Char - width
	}

	for currentLine != nil {
		if lineNum >= offset {
			// render each character in the line separately
			x := 0

			// render line
			for i := skip; i < len(currentLine.Value); i++ {
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
				if isActiveWin {
					if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
						view.SetContent(x, y, string(ch[0]), mcwig.Color("cursor"))
					}
				}

				x += len(ch)
			}

			// render cursor after the end of the line in insert mode
			if lineNum == buf.Cursor.Line && buf.Cursor.Char >= len(currentLine.Value) && isActiveWin {
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
