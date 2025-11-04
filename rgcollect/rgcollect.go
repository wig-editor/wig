package rgcollect

import (
	"fmt"
	"os/exec"

	"github.com/firstrow/wig"
)

func Init(ctx wig.Context) {
	buf := wig.NewBuffer()
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

	cmd := exec.Command("rg", "-n", "KeyHandler")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	wig.TextInsert(buf, buf.Lines.First(), 0, string(output))
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
		StartLine: 1,
		StartChar: 2,
		EndLine:   1,
		EndChar:   7,
	})
	return &wig.HighlighterCursor{
		Cursor: nodes.First(),
	}
}

