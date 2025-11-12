package commands

import (
	"github.com/atotto/clipboard"
	"github.com/firstrow/wig"
)

func CmdClipboardCopy(ctx wig.Context) {
	sel := wig.SelectionToString(ctx.Buf, ctx.Buf.Selection)
	if sel == "" {
		return
	}
	text := wig.SelectionToString(ctx.Buf, ctx.Buf.Selection)
	clipboard.WriteAll(text)
	wig.CmdNormalMode(ctx)
	ctx.Editor.EchoMessage("copy to clipboard")
}

func CmdClipboardPaste(ctx wig.Context) {
	text, _ := clipboard.ReadAll()
	cur := wig.ContextCursorGet(ctx)

	if ctx.Buf.Selection != nil {
		if ctx.Buf.TxStart() {
			if ctx.Buf.Mode() == wig.MODE_VISUAL {
				wig.SelectionDelete(ctx)
			}
			if ctx.Buf.Mode() == wig.MODE_VISUAL_LINE {
				wig.SelectionDelete(ctx)
			}
			line := wig.CursorLine(ctx.Buf, cur)
			wig.TextInsert(ctx.Buf, line, len(line.Value)-1, "\n")
			ctx.Buf.TxEnd()
		}
	}

	wig.TextInsert(ctx.Buf, wig.CursorLine(ctx.Buf, cur), cur.Char, text)
	wig.CmdNormalMode(ctx)
}

