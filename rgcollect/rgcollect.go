package rgcollect

import (
	"fmt"
	"strings"

	"github.com/firstrow/wig"
)

func Init(ctx wig.Context, title string, items []wig.Location) {
	if len(items) == 0 {
		ctx.Editor.EchoMessage("no items found")
		return
	}
	if len(ctx.Editor.Windows) == 1 {
		wig.CmdWindowVSplit(ctx)
	}
	wig.CmdWindowNext(ctx)

	buf := wig.NewBuffer()
	buf.ResetLines()
	buf.FilePath = "[rgcollect " + title + "]"
	buf.Highlighter = &TestHighlighter{}
	buf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_NORMAL: wig.KeyMap{
			"Enter": func(ctx wig.Context) {
				wig.VisitAtLine(ctx, buf, wig.VisitOptions{
					ParseLocation: true,
				})
			},
		},
	})

	ctx.Editor.Buffers = append(ctx.Editor.Buffers, buf)
	ctx.Buf = buf
	wig.EditorInst.ActiveWindow().VisitBuffer(ctx)

	for _, item := range items {
		v := fmt.Sprintf("%s:%d:%d %s", item.FilePath, item.Line, item.Char, strings.TrimSpace(item.Text))
		buf.Append(v)
	}

	wig.CmdWindowNext(ctx)
	wig.VisitAtLine(ctx, buf, wig.VisitOptions{
		Movement:      wig.CmdGotoLine0,
		ParseLocation: true,
	})
}

type TestHighlighter struct{}

func (h *TestHighlighter) Build() {
}

func (h *TestHighlighter) TextChanged(wig.EventTextChange) {
}

func (h *TestHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	return nil
}
