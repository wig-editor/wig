package autocomplete

import (
	"github.com/firstrow/wig"
	"github.com/firstrow/wig/ui"
)

func Register(e *wig.Editor) wig.AutocompleteFn {
	return func(ctx wig.Context) bool {
		items := ctx.Editor.Lsp.Completion(ctx.Buf)

		ui.AutocompleteInit(
			ctx,
			wig.Position{
				Line: ctx.Buf.Cursor.Line,
				Char: ctx.Buf.Cursor.Char,
			},
			items,
		)

		return true
	}
}

