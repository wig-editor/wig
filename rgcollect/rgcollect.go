package rgcollect

import (
	"fmt"
	"strings"

	"github.com/firstrow/wig"
)

func Init(ctx wig.Context, title string, items []wig.Location) {
	if len(ctx.Editor.Windows) == 1 {
		wig.CmdWindowVSplit(ctx)
	}
	wig.CmdWindowNext(ctx)

	buf := wig.NewBuffer()
	buf.ResetLines()
	buf.FilePath = "[rgcollect " + title + "]"
	buf.Highlighter = &TestHighlighter{}
	buf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_INSERT: wig.KeyMap{
			"Enter": func(ctx wig.Context) {
			},
		},
		wig.MODE_NORMAL: wig.KeyMap{
			"Enter": func(ctx wig.Context) {
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
	visitLine(ctx, wig.CmdGotoLine0)
}

type TestHighlighter struct{}

func (h *TestHighlighter) Build() {
}

func (h *TestHighlighter) TextChanged(wig.EventTextChange) {
}

func (h *TestHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	return nil
	// nodes := wig.List[wig.HighlighterNode]{}
	// nodes.PushBack(wig.HighlighterNode{
	// NodeName:  "constant",
	// StartLine: 0,
	// StartChar: 2,
	// EndLine:   0,
	// EndChar:   20,
	// })
	// return &wig.HighlighterCursor{
	// Cursor: nodes.First(),
	// }
}

// Commands
func CmdVisitNextLine(ctx wig.Context) {
	visitLine(ctx, wig.CmdCursorLineDown)
}

func CmdVisitPrevLine(ctx wig.Context) {
	visitLine(ctx, wig.CmdCursorLineUp)
}

func visitLine(ctx wig.Context, upOrDown func(wig.Context)) {
	var rgBuf *wig.Buffer
	var rgWin *wig.Window
	for _, win := range ctx.Editor.Windows {
		if strings.HasPrefix(win.Buffer().FilePath, "[rgcollect") {
			rgBuf = win.Buffer()
			rgWin = win
			break
		}
	}
	if rgBuf == nil {
		ctx.Editor.EchoMessage("rgcollect buffer not visible")
		return
	}

	var line *wig.Element[wig.Line]
	// this is what we need to do
	// to perform action in scope of other window and buffer
	{
		bufCur := wig.WindowCursorGet(rgWin, rgBuf)
		nctx := ctx.Editor.NewContext()
		nctx.Buf = rgBuf
		nctx.Win = rgWin
		upOrDown(nctx)
		line = wig.CursorLine(rgBuf, bufCur)
	}

	filename, lineNum, chNum := wig.ParseFileLocation(line.Value.String(), 0)

	ctx.Buf = ctx.Editor.OpenFile(filename)
	ctx.Editor.ActiveWindow().VisitBuffer(ctx, wig.Cursor{
		Line: lineNum,
		Char: chNum,
	})
}

