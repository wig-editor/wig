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
	hasGitSigns := true
	leftPadding := WindowTextPadding(e, buf)
	signColWidth := 2
	lineNumWidth := 0
	if e.Config.ShowLineNumbers {
		lineNumWidth = len(fmt.Sprintf("%d", buf.CountLines())) + 1
	}
	blameColWidth := 0
	if buf.BlameEnabled && len(buf.BlameLines) > 0 {
		blameColWidth = buf.BlameWidth + 1
	}

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

	// Precalculate visual block bounds for efficient rendering
	var minVisCol, maxVisCol int
	isVisualBlock := buf.Mode() == wig.MODE_VISUAL_BLOCK && buf.Selection != nil
	if isVisualBlock {
		sel := buf.Selection
		startLineNode := wig.CursorLineByNum(buf, sel.Start.Line)
		endLineNode := wig.CursorLineByNum(buf, sel.End.Line)
		startVisCol := wig.VisualCol(startLineNode.Value, sel.Start.Char)
		endVisCol := wig.VisualCol(endLineNode.Value, sel.End.Char)
		minVisCol = min(startVisCol, endVisCol)
		maxVisCol = max(startVisCol, endVisCol)
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

			// Line numbers & Git Signs & Blame
			if e.Config.ShowLineNumbers || hasGitSigns || buf.BlameEnabled {
				xCur := 0
				if hasGitSigns {
					sign, ok := buf.GitSigns[lineNum+1] // GitSigns is 1-indexed
					signStyle := wig.Color("default")
					if xCur >= 0 && xCur < termWidth && y >= 0 && y < termHeight {
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

					if xCur >= 0 && xCur < termWidth && y >= 0 && y < termHeight {
						if lineNum == cur.Line {
							view.SetContent(xCur, y, fmt.Sprintf("%d", lnNum), lineNumTextStyleSelected)
						} else {
							view.SetContent(xCur, y, fmt.Sprintf("%d", lnNum), lineNumTextStyle)
						}
					}
				}
				xCur += lineNumWidth

				if buf.BlameEnabled && blameColWidth > 0 {
					if info, ok := buf.BlameLines[lineNum]; ok {
						if xCur >= 0 && xCur < termWidth && y >= 0 && y < termHeight {
							view.SetContent(xCur, y, info.Display, wig.Color("comment"))
						}
					}
					xCur += blameColWidth
				}
			}
			// End Line Numbers & Git Signs

			// render line
			currVisCol := 0
			for j := 0; j < skip; j++ {
				currVisCol += chlen(currentLine.Value[j])
			}

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
				charLen := chlen(currentLine.Value[i])

				// highlight current line
				if lineNum == cur.Line {
					textStyle = wig.ApplyBg("ui.cursorline", textStyle)
					fillWidth := termWidth - x
					if fillWidth > 0 {
						bg := strings.Repeat(" ", fillWidth)
						view.SetContent(x, y, bg, textStyle)
					}
				}

				// selection
				if buf.Selection != nil {
					if isVisualBlock {
						sel := buf.Selection
						minLine, maxLine := sel.Start.Line, sel.End.Line
						if minLine > maxLine {
							minLine, maxLine = maxLine, minLine
						}
						if lineNum >= minLine && lineNum <= maxLine {
							// Highlight based on visual screen columns
							if currVisCol < maxVisCol && currVisCol+charLen > minVisCol {
								textStyle = wig.ApplyBg("ui.selection.primary", textStyle)
							}
						}
					} else if wig.SelectionCursorInRange(buf.Selection, wig.Cursor{Line: lineNum, Char: i}) {
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

				if i > len(currentLine.Value)-1 {
					continue
				}

				ch := getRenderChar(currentLine.Value[i])

				// todo: handle tabs colors?
				// render text
				if x >= 0 && x < termWidth && y >= 0 && y < termHeight {
					view.SetContent(x, y, string(ch), textStyle)
				}

				// render cursor
				if isActiveWin {
					if lineNum == cur.Line && i == cur.Char {
						baseCursor := wig.Color("default")
						if c, found := wig.FindColor("ui.selection"); found {
							baseCursor = c
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
						if x >= 0 && x < termWidth && y >= 0 && y < termHeight {
							view.SetContent(x, y, string(ch[0]), baseCursor)
						}
					}
				}

				x += charLen
				currVisCol += charLen
			}

			// render cursor after the end of the line in insert mode
			if lineNum == cur.Line && cur.Char >= len(currentLine.Value) && isActiveWin {
				if x >= 0 && x < termWidth && y >= 0 && y < termHeight {
					view.SetContent(x, y, " ", wig.Color("ui.cursor"))
				}
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

// WindowTextPadding returns the X offset at which buffer text begins inside
// a window's viewport, accounting for the git sign column, line-number
// gutter, and optional blame column.
//
// Every renderer (text in WindowRender, indent guides in
// render.RenderIndentGuides, future overlays like diagnostics / virtual
// text) MUST use this function. Hardcoding the layout in two places is
// what previously caused indent guides to drift when the git-sign column
// was always reserved by WindowRender (hasGitSigns := true) but
// RenderIndentGuides still used len(buf.GitSigns) > 0.
func WindowTextPadding(e *wig.Editor, buf *wig.Buffer) int {
	// Always reserve the sign column so the gutter width is stable
	// whether or not the current buffer currently has git signs. This
	// matches WindowRender's hasGitSigns := true behaviour.
	signColWidth := 2

	lineNumWidth := 0
	if e.Config.ShowLineNumbers {
		lineNumWidth = len(fmt.Sprintf("%d", buf.CountLines())) + 1
	}

	blameColWidth := 0
	if buf.BlameEnabled && len(buf.BlameLines) > 0 {
		blameColWidth = buf.BlameWidth + 1
	}

	return signColWidth + lineNumWidth + blameColWidth
}
