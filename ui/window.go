package ui

import (
	"fmt"
	"strings"

	str "github.com/boyter/go-string"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"

	"github.com/firstrow/wig"
)

func nodeToColor(node *wig.Element[wig.HighlighterNode]) tcell.Style {
	if node == nil {
		return wig.Color("default")
	}

	return wig.Color(node.Value.NodeName)
}

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

	// Line numbers config
	lineNum := 0
	lineNumTextStyle := wig.Color("ui.linenr")
	lineNumTextStyleSelected := wig.Color("ui.linenr.selected")
	hasGitSigns := len(buf.GitSigns) > 0
	leftPadding := 0
	signColWidth := 0
	if hasGitSigns {
		signColWidth = 2
	}
	lineNumWidth := 0
	if e.Config.ShowLineNumbers {
		lineNumWidth = len(fmt.Sprintf("%d", buf.CountLines())) + 1
	}
	leftPadding = signColWidth + lineNumWidth

	y := 0

	isActiveWin := win == e.ActiveWindow()

	skip := 0
	if cur.Char > termWidth {
		skip = cur.Char - termWidth
	}

	startLine := uint32(offset)
	var tsNodeCursor *wig.HighlighterCursor
	if buf.Highlighter != nil {
		// TODO: query new highlights only if visible are have changed.
		// Now it reloads colors on j,k,l, basically on any key movement.
		tsNodeCursor = buf.Highlighter.ForRange(uint32(startLine), startLine+uint32(termHeight))
	}

	for currentLine != nil {
		if lineNum >= offset && y <= termHeight {
			// render each character in the line separately
			x := leftPadding // onscreen position

			// highlight search
			searchMatches := [][]int{}
			if wig.LastSearchPattern != "" {
				searchMatches = str.IndexAllIgnoreCase(string(currentLine.Value), wig.LastSearchPattern, -1)
			}

			diagnostics := e.Lsp.Diagnostics(buf, lineNum)

			// Line numbers & Git Signs
			if e.Config.ShowLineNumbers || hasGitSigns {
				xCur := 0
				if hasGitSigns {
					sign, ok := buf.GitSigns[lineNum+1] // GitSigns is 1-indexed
					signStyle := wig.Color("default")
					if ok {
						if sign == '+' {
							signStyle = wig.Color("diff.plus")
						} else if sign == '-' {
							signStyle = wig.Color("diff.minus")
						} else if sign == '~' {
							signStyle = wig.Color("diff.delta")
						}
						view.SetContent(xCur, y, string(sign), signStyle)
					} else {
						view.SetContent(xCur, y, " ", signStyle)
					}
					xCur += signColWidth
				}

				if e.Config.ShowLineNumbers {
					lnNum := lineNum + 1
					if e.Config.RelativeLineNumbers {
						lnNum = cur.Line - lineNum
						if lnNum < 0 {
							lnNum = -lnNum
						}
						// If hybrid mode is enabled, show absolute line number on current line
						if e.Config.CurrentLineAbsolute && lineNum == cur.Line {
							lnNum = lineNum + 1
						}
					}

					if lineNum == cur.Line {
						view.SetContent(xCur, y, fmt.Sprintf("%d", lnNum), lineNumTextStyleSelected)
					} else {
						view.SetContent(xCur, y, fmt.Sprintf("%d", lnNum), lineNumTextStyle)
					}
				}
			}
			// End Line Numbers & Git Signs

			// render line
			for i := skip; i < len(currentLine.Value); i++ {
				// render selection
				textStyle := wig.Color("default")

				if tsNodeCursor != nil {
					colorNode, ok := tsNodeCursor.Seek(uint32(lineNum), uint32(i))
					if ok {
						textStyle = nodeToColor(colorNode)
					}
				}

				// Colors and styles

				// highlight current line
				if lineNum == cur.Line {
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
						// once character error
						if info.Range.Start.Character == info.Range.End.Character && info.Range.End.Character == uint32(i) {
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
