package mcwig

import (
	"context"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"
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
	e          *Editor
	buf        *Buffer
	nodes      List[TreeSitterRangeNode]
	parser     *sitter.Parser
	q          *sitter.Query
	tree       *sitter.Tree
	sourceCode []byte
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
		h.e.LogError(errors.Wrap(err, "tree sitter query error"))
		return
	}

	h.Build()
	buf.Highlighter = h
}

// Build and "syntax" must be separated.
// so we can do full rebuild and query syntax ranges only
// for lines on the screen.
func (h *Highlighter) Build() {
	h.nodes = List[TreeSitterRangeNode]{}

	if h.tree != nil {
		h.tree.Close()
	}

	h.sourceCode = []byte(h.buf.String())
	tree, err := h.parser.ParseCtx(context.Background(), nil, h.sourceCode)
	if err != nil {
		panic(err.Error())
	}

	h.tree = tree
}

func (h *Highlighter) RootNode() *Element[TreeSitterRangeNode] {
	return h.nodes.First()
}

func (h *Highlighter) Highlights(lineStart, lineEnd uint32) {
	h.nodes = List[TreeSitterRangeNode]{}

	qc := sitter.NewQueryCursor()
	qc.SetPointRange(
		sitter.Point{Row: lineStart, Column: 0},
		sitter.Point{Row: lineEnd, Column: 0},
	)
	qc.Exec(h.q, h.tree.RootNode())

	defer qc.Close()

	i := 0

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		// Apply predicates filtering
		m = qc.FilterPredicates(m, h.sourceCode)
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
		i++
	}
}

func NodeToColor(node *Element[TreeSitterRangeNode]) tcell.Style {
	if node == nil {
		return Color("default")
	}

	return Color(node.Value.NodeName)
}

type TreeSitterNodeCursor struct {
	cursor *Element[TreeSitterRangeNode]
}

func NewColorNodeCursor(rootNode *Element[TreeSitterRangeNode]) *TreeSitterNodeCursor {
	if rootNode == nil {
		return nil
	}
	return &TreeSitterNodeCursor{
		cursor: rootNode,
	}
}

func (c *TreeSitterNodeCursor) Seek(line, ch uint32) (node *Element[TreeSitterRangeNode], found bool) {
	inRange := func(node *Element[TreeSitterRangeNode], line, ch uint32) bool {
		if node == nil {
			return false
		}
		if line >= node.Value.StartLine && line <= node.Value.EndLine {
			if ch >= node.Value.StartChar && ch < node.Value.EndChar {
				return true
			}
		}
		return false
	}

	if inRange(c.cursor, line, ch) {
		return c.cursor, true
	}

	nextNode := c.cursor.Next()
	for nextNode != nil {
		if nextNode.Value.StartLine > line {
			break
		}

		if inRange(nextNode, line, ch) {
			c.cursor = nextNode
			return c.cursor, true
		}

		nextNode = nextNode.Next()
	}

	return nil, false
}
