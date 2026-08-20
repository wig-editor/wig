package wig

import "github.com/gdamore/tcell/v2"

// indentGuideTabWidth must match the tab width used in VisualCol (cursor.go).
const indentGuideTabWidth = 4

// IndentGuideColumns returns the visual column positions (0-indexed from
// the start of the text content) where vertical indent guide lines should
// be drawn for the given line content.
//
// A guide is placed at the first visual column of each tab-stop within the
// line's leading whitespace — one guide per indent level. For example, a
// line indented with two tabs (tab width 4) produces guides at visual
// columns 0 and 4 (i.e. "|\t|\tcode").
//
// Lines with less than one tab-stop of leading whitespace produce no guides.
func IndentGuideColumns(lineRun []rune) []int {
	if len(lineRun) == 0 {
		return nil
	}
	visualCol := 0
	for _, r := range lineRun {
		if r == '\t' {
			visualCol += indentGuideTabWidth - (visualCol % indentGuideTabWidth)
		} else if r == ' ' {
			visualCol++
		} else {
			break
		}
	}
	if visualCol < indentGuideTabWidth {
		return nil
	}
	var positions []int
	for col := 0; col+indentGuideTabWidth <= visualCol; col += indentGuideTabWidth {
		positions = append(positions, col)
	}
	return positions
}

// IndentGuideSet returns a map of visual columns that have indent guides,
// suitable for O(1) lookup during per-character rendering. Returns nil if
// the line has no indent guides.
func IndentGuideSet(lineRun []rune) map[int]bool {
	positions := IndentGuideColumns(lineRun)
	if len(positions) == 0 {
		return nil
	}
	set := make(map[int]bool, len(positions))
	for _, p := range positions {
		set[p] = true
	}
	return set
}

// IndentGuideStyle returns the tcell style for indent guide lines.
// Configure the color via the "ui.indentguide" key in your theme.
func IndentGuideStyle() tcell.Style {
	return Color("ui.indentguide")
}

// IndentGuideGlyph is the character used to draw vertical indent guides.
const IndentGuideGlyph = "│"

// RenderIndentGuidesForLine draws indent guide characters for a single
// rendered line. Call this AFTER the line's text content has been rendered,
// as it overwrites whitespace cells with the guide glyph.
//
// style is supplied by the caller rather than computed here so that lines
// with a background already painted over them (e.g. the cursorline) can
// pass a blended style (see ApplyBg("ui.cursorline", IndentGuideStyle())).
// Using a flat, backgroundless style unconditionally would reset those
// cells back to the default background, punching a solid block out of
// the cursorline highlight instead of drawing a thin guide over it.
//
// Parameters:
//   - view:      the render target
//   - lineRun:   the rune content of the line being rendered
//   - textX:     screen X where the first visible text column is drawn
//   - scrollX:   horizontal scroll offset in visual columns (0 = no scroll)
//   - screenY:   screen Y of this line
//   - viewWidth: number of columns available for text
//   - style:     the style to draw the guide glyph with (see IndentGuideStyle)
func RenderIndentGuidesForLine(view View, lineRun []rune, textX, scrollX, screenY, viewWidth int, style tcell.Style) {
	positions := IndentGuideColumns(lineRun)
	if len(positions) == 0 {
		return
	}
	for _, pos := range positions {
		relX := pos - scrollX
		if relX < 0 || relX >= viewWidth {
			continue
		}
		screenX := textX + relX
		view.SetContent(screenX, screenY, IndentGuideGlyph, style)
	}
}
