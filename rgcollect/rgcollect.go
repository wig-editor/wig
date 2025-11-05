package rgcollect

import (
	"fmt"
	"strings"

	"github.com/firstrow/wig"
)

func Init(ctx wig.Context, items []wig.Location) {
	buf := wig.NewBuffer()
	buf.ResetLines()
	buf.FilePath = "[rgcollect]"
	ctx.Editor.Buffers = append(ctx.Editor.Buffers, buf)
	wig.EditorInst.ActiveWindow().ShowBuffer(buf)

	buf.Highlighter = &TestHighlighter{}

	buf.KeyHandler = wig.DefaultKeyHandler(wig.ModeKeyMap{
		wig.MODE_INSERT: wig.KeyMap{
			"Enter": func(ctx wig.Context) {
				fmt.Println("1111111111")
			},
		},
		wig.MODE_NORMAL: wig.KeyMap{
			"Enter": func(ctx wig.Context) {
				fmt.Println("2222222222222")
			},
		},
	})

	for _, item := range items {
		v := fmt.Sprintf("%s:(%d:%d) %s", item.FilePath, item.Line, item.Char, strings.TrimSpace(item.Text))
		buf.Append(v)
	}
}

type TestHighlighter struct{}

func (h *TestHighlighter) Build() {
}

func (h *TestHighlighter) TextChanged(wig.EventTextChange) {
}

func (h *TestHighlighter) ForRange(startLine, endLine uint32) *wig.HighlighterCursor {
	nodes := wig.List[wig.HighlighterNode]{}
	nodes.PushBack(wig.HighlighterNode{
		NodeName:  "constant",
		StartLine: 0,
		StartChar: 2,
		EndLine:   0,
		EndChar:   2,
	})
	return &wig.HighlighterCursor{
		Cursor: nodes.First(),
	}
}

