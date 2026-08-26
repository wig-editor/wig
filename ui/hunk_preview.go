package ui

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

// HunkPreviewWidget is a lightweight, filter-less popup for previewing text
// blocks (like git hunks or LSP hover). Unlike UiPicker it captures no text
// input, so every key is available as a one-key shortcut (e.g. "r" to revert).
type HunkPreviewWidget struct {
	e            *wig.Editor
	keymap       *wig.KeyHandler
	header       string
	lines        []string
	scrollOffset int
	onRevert     func(wig.Context)
	lineStyler   func(string) tcell.Style
	posX         int
	posY         int
}

func (u *HunkPreviewWidget) Plane() wig.RenderPlane {
	return wig.PlaneEditor
}

// expandTabs replaces tab characters with spaces so on-screen column math
// (truncate, x offsets, box clearing) stays in sync with what the terminal
// actually draws. Raw diff lines from git preserve source indentation
// (often tabs), and rendering those verbatim desyncs our column tracking
// from the terminal's own tab-stop rendering, producing garbled overlap.
func expandTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

// HunkPreviewInit opens a read-only popup showing header+lines (as produced
// by a unified diff hunk or LSP hover). onRevert, if non-nil, is invoked
// (with the popup's own context) when the user presses "r"; the popup then closes.
// lineStyler allows customizing the color of each line (e.g. diff colors).
func HunkPreviewInit(ctx wig.Context, header string, lines []string, onRevert func(wig.Context), lineStyler func(string) tcell.Style, pos ...int) *HunkPreviewWidget {
	cleanLines := make([]string, len(lines))
	for i, l := range lines {
		cleanLines[i] = expandTabs(strings.TrimRightFunc(l, unicode.IsSpace))
	}

	if lineStyler == nil {
		lineStyler = lineStyle
	}

	var posX, posY int = -1, -1
	if len(pos) >= 2 {
		posX = pos[0]
		posY = pos[1]
	}

	widget := &HunkPreviewWidget{
		e:          ctx.Editor,
		header:     expandTabs(header),
		lines:      cleanLines,
		onRevert:   onRevert,
		lineStyler: lineStyler,
		posX:       posX,
		posY:       posY,
	}

	km := wig.KeyMap{
		"Esc": func(ctx wig.Context) {
			ctx.Editor.PopUi()
		},
		"q": func(ctx wig.Context) {
			ctx.Editor.PopUi()
		},
		"j": func(ctx wig.Context) {
			widget.scrollDown(1)
		},
		"k": func(ctx wig.Context) {
			widget.scrollUp(1)
		},
		"Down": func(ctx wig.Context) {
			widget.scrollDown(1)
		},
		"Up": func(ctx wig.Context) {
			widget.scrollUp(1)
		},
		"ctrl+d": func(ctx wig.Context) {
			widget.scrollDown(10)
		},
		"ctrl+u": func(ctx wig.Context) {
			widget.scrollUp(10)
		},
	}

	if widget.onRevert != nil {
		km["r"] = func(ctx wig.Context) {
			defer ctx.Editor.PopUi()
			widget.onRevert(ctx)
		}
	}

	widget.keymap = wig.NewKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: km,
	})

	ctx.Editor.PushUi(widget)
	return widget
}

func (u *HunkPreviewWidget) scrollDown(n int) {
	maxOffset := max(len(u.lines)-1, 0)
	u.scrollOffset = min(u.scrollOffset+n, maxOffset)
}

func (u *HunkPreviewWidget) scrollUp(n int) {
	u.scrollOffset = max(u.scrollOffset-n, 0)
}

func (u *HunkPreviewWidget) Mode() wig.Mode {
	return wig.MODE_NORMAL
}

func (u *HunkPreviewWidget) Keymap() *wig.KeyHandler {
	return u.keymap
}

func lineStyle(line string) tcell.Style {
	if len(line) == 0 {
		return wig.Color("default")
	}
	switch line[0] {
	case '+':
		return wig.Color("diff.plus")
	case '-':
		return wig.Color("diff.minus")
	default:
		return wig.Color("default")
	}
}

func (u *HunkPreviewWidget) Render(view wig.View) {
	vw, vh := view.Size()

	// Calculate dynamic width based on content
	maxLineWidth := 0
	for _, l := range u.lines {
		w := len([]rune(l))
		if w > maxLineWidth {
			maxLineWidth = w
		}
	}

	w := maxLineWidth + 4
	h := len(u.lines) + 4

	// Clamp dimensions to screen
	if w > int(float32(vw)*0.8) {
		w = int(float32(vw) * 0.8)
	}
	if h > vh-4 {
		h = vh - 4
	}
	if w < 20 {
		w = 20
	}
	if h < 5 {
		h = 5
	}

	var x, y int
	if u.posX >= 0 || u.posY >= 0 {
		// Follow cursor position
		x = u.posX
		// Place popup below the cursor line
		y = u.posY + 1

		// Adjust if overflowing right edge
		if x+w > vw {
			x = vw - w
		}
		if x < 0 {
			x = 0
		}

		// If overflowing bottom edge, try placing exactly above the cursor
		if y+h > vh {
			y = u.posY - h
		}

		// Clamp to screen bounds
		if y < 0 {
			y = 0
		}
		if y+h > vh {
			y = vh - h
			if y < 0 {
				y = 0
			}
		}
	} else {
		// Centered layout
		x = vw/2 - w/2
		y = 3
	}

	boxStyle := wig.Color("ui.menu")
	drawBox2(view, x, y, w, h, boxStyle)

	// header as title on the border
	if u.header != "" {
		titleStr := " " + truncate(u.header, w-4) + " "
		view.SetContent(x+2, y, titleStr, wig.Color("ui.linenr"))
	}

	// separator
	sep := strings.Repeat(string('─'), w-3)
	view.SetContent(x+2, y+1, sep, boxStyle)

	// body (scrollable)
	pageSize := h - 4 // rows available below separator, above hint line
	if pageSize < 1 {
		pageSize = 1
	}

	start := u.scrollOffset
	if start > len(u.lines) {
		start = len(u.lines)
	}
	end := start + pageSize
	if end > len(u.lines) {
		end = len(u.lines)
	}

	for i, line := range u.lines[start:end] {
		view.SetContent(x+2, y+2+i, truncate(line, w-4), u.lineStyler(line))
	}

	// hint line
	hint := "[r] revert  [Esc/q] close"
	if u.onRevert == nil {
		hint = "[Esc/q] close"
	}
	view.SetContent(x+2, y+h-1, fmt.Sprintf("%s", hint), wig.Color("ui.linenr"))
}
