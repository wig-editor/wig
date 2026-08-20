package ui

import (
	"strings"
	"unicode"

	"github.com/firstrow/wig"
	"github.com/gdamore/tcell/v2"
)

// HoverInit displays LSP hover responses in a temporary centered popup box.
// It leverages HunkPreviewWidget to share scrolling and rendering logic.
func HoverInit(ctx wig.Context, text string) *HunkPreviewWidget {
	cleanLines := make([]string, 0, 16)
	for _, l := range strings.Split(text, "\n") {
		cleanLines = append(cleanLines, strings.TrimRightFunc(l, unicode.IsSpace))
	}

	// Use a plain styler so hover text isn't colored like a diff
	plainStyler := func(line string) tcell.Style {
		return wig.Color("default")
	}

	return HunkPreviewInit(ctx, "Hover", cleanLines, nil, plainStyler)
}
