package mcwig

import (
	"context"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

type TreeSitterRangeNode struct {
	NodeName  string
	StartLine uint32
	StartChar uint32
	EndLine   uint32
	EndChar   uint32
}

type Highlighter struct {
	e     *Editor
	buf   *Buffer
	nodes List[TreeSitterRangeNode]
}

func HighlighterInitBuffer(e *Editor, buf *Buffer) {
	if !strings.HasSuffix(buf.FilePath, ".go") {
		return
	}

	h := &Highlighter{
		e:     e,
		buf:   buf,
		nodes: List[TreeSitterRangeNode]{},
	}
	h.Build()
	buf.Highlighter = h
}

// on file open should be called!
// builds list of nodes with color identifiers and ranges
func (h *Highlighter) Build() {
	h.nodes = List[TreeSitterRangeNode]{}

	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	sourceCode := []byte(h.buf.String())

	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		panic(err.Error())
	}

	hgFile := "/home/andrew/code/mcwig/runtime/helix/go/highlights.scm"
	highlightQ, _ := os.ReadFile(hgFile)
	q, err := sitter.NewQuery(highlightQ, golang.GetLanguage())
	if err != nil {
		h.e.LogError(err)
		return
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		for _, c := range m.Captures {
			// fmt.Println(q.CaptureNameForId(c.Index))
			// fmt.Println(">", c.Node.Type(), c.Node.StartPoint(), c.Node.EndPoint())
			node := TreeSitterRangeNode{
				NodeName:  q.CaptureNameForId(c.Index),
				StartLine: c.Node.StartPoint().Row,
				StartChar: c.Node.StartPoint().Column,
				EndLine:   c.Node.EndPoint().Row,
				EndChar:   c.Node.EndPoint().Column,
			}
			h.nodes.PushBack(node)
		}
	}
}

func (h *Highlighter) RootNode() *Element[TreeSitterRangeNode] {
	return h.nodes.First()
}

func GetColorNode(node *Element[TreeSitterRangeNode], line uint32, ch uint32) *Element[TreeSitterRangeNode] {
	if node == nil {
		return nil
	}

	if line >= node.Value.StartLine && line <= node.Value.EndLine {
		if ch >= node.Value.StartChar && ch < node.Value.EndChar {
			return node
		}
	}

	return GetColorNode(node.Next(), line, ch)
}

func NodeToColor(node *Element[TreeSitterRangeNode]) tcell.Style {
	if node == nil {
		return Color("default")
	}

	return Color(node.Value.NodeName)
}
