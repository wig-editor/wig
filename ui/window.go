package ui

import (
	str "github.com/boyter/go-string"
	"github.com/firstrow/mcwig"
	"github.com/mattn/go-runewidth"
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

	var tsNodeCursor *mcwig.TreeSitterNodeCursor
	if buf.Highlighter != nil {
		startLine := uint32(0)
		if offset > 0 {
			startLine = uint32(offset)
		}
		buf.Highlighter.Highlights(uint32(startLine), startLine+uint32(termHeight))
		tsNodeCursor = mcwig.NewColorNodeCursor(buf.Highlighter.RootNode())
	}

	for currentLine != nil {
		if lineNum >= offset && y <= termHeight {
			// render each character in the line separately
			x := 0 // onscreen position

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

				// TODO: search highlight and selection check must be moved out of the loop
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

				if tsNodeCursor != nil {
					colorNode, ok := tsNodeCursor.Seek(uint32(lineNum), uint32(i))
					if ok && tempColor == 0 {
						textStyle = mcwig.NodeToColor(colorNode)
					}
				}

				// todo: handle tabs colors?
				view.SetContent(x, y, string(ch), textStyle)

				// render cursor
				if isActiveWin {
					if lineNum == buf.Cursor.Line && i == buf.Cursor.Char {
						view.SetContent(x, y, string(ch[0]), mcwig.Color("ui.cursor"))
					}
				}

				x += chlen(currentLine.Value[i])
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

func chlen(c rune) int {
	if c == '\t' {
		return 4
	}
	if c == '\n' {
		return 0
	}
	return runewidth.RuneWidth(c)
}

func getRenderChar(c rune) string {
	if c == '\t' {
		return "    "
	}
	if c == '\n' {
		return " "
	}
	return string(c)
}

