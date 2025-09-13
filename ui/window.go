package ui

import (
	"strings"

	str "github.com/boyter/go-string"
	"github.com/mattn/go-runewidth"

	"github.com/firstrow/wig"
)

func WindowRender(e *wig.Editor, view wig.View, win *wig.Window) {
	buf := win.Buffer()
	if buf == nil {
		return
	}
	cur := wig.WindowCursorGet(win, buf)

	termWidth, termHeight := view.Size()
	termWidth -= 2
	termHeight -= 1

	currentLine := buf.Lines.First()
	offset := cur.ScrollOffset

	lineNum := 0
	y := 0

	isActiveWin := win == e.ActiveWindow()

	skip := 0
	if cur.Char > termWidth {
		skip = cur.Char - termWidth
	}

	var tsNodeCursor *wig.TreeSitterNodeCursor
	if buf.Highlighter != nil {
		startLine := uint32(0)
		if offset > 0 {
			startLine = uint32(offset)
		}
		buf.Highlighter.Highlights(uint32(startLine), startLine+uint32(termHeight))
		tsNodeCursor = wig.NewColorNodeCursor(buf.Highlighter.RootNode())
	}

	for currentLine != nil {
		if lineNum >= offset && y <= termHeight {
			// render each character in the line separately
			x := 0 // onscreen position

			// highlight search
			searchMatches := [][]int{}
			if wig.LastSearchPattern != "" {
				searchMatches = str.IndexAllIgnoreCase(string(currentLine.Value), wig.LastSearchPattern, -1)
			}

			diagnostics := e.Lsp.Diagnostics(buf, lineNum)

			// render line
			for i := skip; i < len(currentLine.Value); i++ {
				// render selection
				textStyle := wig.Color("default")

				if tsNodeCursor != nil {
					colorNode, ok := tsNodeCursor.Seek(uint32(lineNum), uint32(i))
					if ok {
						textStyle = wig.NodeToColor(colorNode)
					}
				}

				// Colors and styles

				// highlight current line
				if lineNum == cur.Line && isActiveWin {
					textStyle = wig.ApplyBg("ui.cursorline", textStyle)
					bg := strings.Repeat(" ", termWidth)
					view.SetContent(x, y, bg, textStyle)
				}

				// selection
				if buf.Selection != nil {
					if wig.SelectionCursorInRange(buf.Selection, wig.Cursor{Line: lineNum, Char: i}) {
						textStyle = wig.ApplyBg("ui.selection.primary", textStyle)
					}
				}

				// highlight search
				if len(searchMatches) > 0 {
					for _, m := range searchMatches {
						if i >= m[0] && i < m[1] {
							textStyle = wig.ApplyBg("ui.selection", textStyle)
						}
					}
				}

				// lsp errors
				if len(diagnostics) > 0 {
					for _, info := range diagnostics {
						if i >= int(info.Range.Start.Character) && i < int(info.Range.End.Character) {
							textStyle = wig.MergeStyles(textStyle, "diagnostic.error")
						}
					}
				}

				/////////////////////////////////

				ch := getRenderChar(currentLine.Value[i])

				// todo: handle tabs colors?
				// render text
				view.SetContent(x, y, string(ch), textStyle)

				// render cursor
				if isActiveWin {
					if lineNum == cur.Line && i == cur.Char {
						baseCursor, found := wig.FindColor("ui.selection")
						if !found {
							panic("theme ui.selection not defined")
						}
						if c, found := wig.FindColor("ui.cursor"); found {
							baseCursor = c
						}
						if buf.Mode() == wig.MODE_INSERT {
							if c, found := wig.FindColor("ui.cursor.primary.insert"); found {
								baseCursor = c
							}
						}
						if buf.Mode() == wig.MODE_VISUAL {
							if c, found := wig.FindColor("ui.cursor.primary.select"); found {
								baseCursor = c
							}
						}
						view.SetContent(x, y, string(ch[0]), baseCursor)
					}
				}

				x += chlen(currentLine.Value[i])
			}

			// render cursor after the end of the line in insert mode
			if lineNum == cur.Line && cur.Char >= len(currentLine.Value) && isActiveWin {
				view.SetContent(x, y, " ", wig.Color("ui.cursor"))
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

