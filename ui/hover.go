package ui

import (
	"strings"
	"unicode"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

// HoverInit displays LSP hover responses in a popup box positioned under the cursor.
// It leverages HunkPreviewWidget to share scrolling and rendering logic.
func HoverInit(ctx wig.Context, text string) *HunkPreviewWidget {
	cleanLines := make([]string, 0, 16)
	for _, l := range strings.Split(text, "\n") {
		cleanLines = append(cleanLines, strings.TrimRightFunc(l, unicode.IsSpace))
	}

	// Use a styler that matches the popup's background for better visual integration
	plainStyler := func(line string) tcell.Style {
		return wig.ApplyBg("ui.menu", wig.Color("default"))
	}

	// Calculate cursor screen position for popup placement
	cur := wig.ContextCursorGet(ctx)
	line := wig.CursorLine(ctx.Buf, cur)
	visCol := wig.VisualCol(line.Value, cur.Char)
	posX := visCol + WindowTextPadding(ctx.Editor, ctx.Buf)
	posY := cur.Line - cur.ScrollOffset

	return HunkPreviewInit(ctx, "Hover", cleanLines, nil, plainStyler, posX, posY)
}
