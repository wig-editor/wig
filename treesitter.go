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
	e      *Editor
	buf    *Buffer
	nodes  List[TreeSitterRangeNode]
	parser *sitter.Parser
	q      *sitter.Query
}

// TODO: highlighter parser and query MUST be request
// for every buffer
func HighlighterInitBuffer(e *Editor, buf *Buffer) {
	if !strings.HasSuffix(buf.FilePath, ".go") {
		return
	}

	h := &Highlighter{
		e:     e,
		buf:   buf,
		nodes: List[TreeSitterRangeNode]{},
	}
	h.parser = sitter.NewParser()
	h.parser.SetLanguage(golang.GetLanguage())

	hgFile := "/home/andrew/code/mcwig/runtime/helix/go/highlights.scm"
	highlightQ, _ := os.ReadFile(hgFile)

	var err error
	h.q, err = sitter.NewQuery(highlightQ, golang.GetLanguage())
	if err != nil {
		h.e.LogError(err)
		return
	}

	h.Build()
	buf.Highlighter = h
}

func (h *Highlighter) Build() {
	h.nodes = List[TreeSitterRangeNode]{}

	sourceCode := []byte(h.buf.String())
	tree, err := h.parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil {
		panic(err.Error())
	}
	defer tree.Close()

	qc := sitter.NewQueryCursor()
	qc.Exec(h.q, tree.RootNode())
	defer qc.Close()

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		for _, c := range m.Captures {
			startPoint := c.Node.StartPoint()
			endPoint := c.Node.EndPoint()
			h.nodes.PushBack(TreeSitterRangeNode{
				NodeName:  h.q.CaptureNameForId(c.Index),
				StartLine: startPoint.Row,
				StartChar: startPoint.Column,
				EndLine:   endPoint.Row,
				EndChar:   endPoint.Column,
			})
		}
	}
}

func (h *Highlighter) RootNode() *Element[TreeSitterRangeNode] {
	return h.nodes.First()
}

func GetColorNode(node *Element[TreeSitterRangeNode], line uint32, ch uint32) *Element[TreeSitterRangeNode] {
	return nil

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
