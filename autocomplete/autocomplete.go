package autocomplete

import (
	"github.com/firstrow/wig"
	"github.com/firstrow/wig/ui"
)

func Register(e *wig.Editor) wig.AutocompleteFn {
	return func(ctx wig.Context) bool {
		// Check for snippets first
		if ctx.Editor.Snippets.Complete(ctx) {
			return true
		}

		// Lsp completion
		cur := wig.ContextCursorGet(ctx)
		items := ctx.Editor.Lsp.Completion(ctx.Buf)
		ui.AutocompleteInit(
			ctx,
			wig.Position{
				Line: cur.Line,
				Char: cur.Char,
			},
			items,
		)

		return true
	}
}

