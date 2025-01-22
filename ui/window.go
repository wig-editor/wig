package ui

import (
	str "github.com/boyter/go-string"
	"github.com/firstrow/mcwig"
)

func WindowRender(e *mcwig.Editor, view mcwig.View, win *mcwig.Window) {
	buf := win.Buffer()
	if buf == nil {
		return
	}

	termWidth, termHeight := view.Size()
	termWidth -= 2
	termHeight -= 1

	currentLine := buf.Lines.First()
	offset := buf.ScrollOffset
	lineNum := 0
	y := 0

	isActiveWin := win == e.ActiveWindow()

	skip := 0
	if buf.Cursor.Char > termWidth {
		skip = buf.Cursor.Char - termWidth
	}

	colorNode := buf.Highlighter.RootNode()

	for currentLine != nil {
		if lineNum >= offset && y <= termHeight {
			// render each character in the line separately
			x := 0

			// highlight search
			searchMatches := [][]int{}
			if mcwig.LastSearchPattern != "" {
				searchMatches = str.IndexAllIgnoreCase(string(currentLine.Value), mcwig.LastSearchPattern, -1)
			}

			// render line
			for i := skip; i < len(currentLine.Value); i++ {
				// render selection
				textStyle := mcwig.Color("default")
				tempColor := 0

				// highlight search
				if len(searchMatches) > 0 {
					for _, m := range searchMatches {
						if i >= m[0] && i < m[1] {
							textStyle = mcwig.Color("ui.cursor.match")
							tempColor = 1
						}
					}
				}

				// selection
				if buf.Selection != nil {
					if mcwig.SelectionCursorInRange(buf.Selection, mcwig.Cursor{Line: lineNum, Char: i}) {
						textStyle = mcwig.Color("ui.selection")
						tempColor = 1
					}
				}

				ch := getRenderChar(currentLine.Value[i])
				colorNode := mcwig.GetColorNode(colorNode, uint32(lineNum), uint32(i))
				if tempColor == 0 {
					textStyle = mcwig.NodeToColor(colorNode)
				}

				// todo: handle tabs colors?
				view.SetContent(x, y, string(ch), textStyle)

				// render cursor
				if isActiveWin {
					if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
						view.SetContent(x, y, string(ch[0]), mcwig.Color("ui.cursor"))
					}
				}

				x += len(ch)
			}

			// render cursor after the end of the line in insert mode
			if lineNum == buf.Cursor.Line && buf.Cursor.Char >= len(currentLine.Value) && isActiveWin {
				view.SetContent(x, y, " ", mcwig.Color("ui.cursor"))
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
