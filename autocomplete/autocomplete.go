package autocomplete

import (
	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/ui"
)

func Register(e *mcwig.Editor) mcwig.AutocompleteFn {
	return func(ctx mcwig.Context) bool {
		items := ctx.Editor.Lsp.Completion(ctx.Buf)

		ui.AutocompleteInit(
			ctx,
			mcwig.Position{
				Line: ctx.Buf.Cursor.Line,
				Char: ctx.Buf.Cursor.Char,
			},
			items,
		)

		return true
	}
}

