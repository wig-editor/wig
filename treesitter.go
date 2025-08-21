package wig

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// TODO: rewrite treesitter to use channel and scheduled parsing
var tslock sync.Mutex

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

// TODO: this must exit on editor close
// use context cancel()
func HighlighterGo(e *Editor) {
	go func() {
		events := e.Events.Subscribe()
		for {
			select {
			case event := <-events:
				switch e := event.Msg.(type) {
				case EventTextChange:
					HighlighterEditTree(e)
				}
				event.Wg.Done()
			}
		}
	}()
}

func HighlighterEditTree(event EventTextChange) {
	if event.Buf == nil {
		return
	}

	h := event.Buf.Highlighter
	if h == nil {
		return
	}

	tslock.Lock()
	defer tslock.Unlock()

	ll := HighlighterAdaptEditInput(event)
	event.Buf.Highlighter.tree.Edit(ll)

	h.nodes = List[TreeSitterRangeNode]{}
	event.Buf.Highlighter.sourceCode = []byte(event.Buf.String())
	tree, err := h.parser.ParseCtx(context.Background(), h.tree, event.Buf.Highlighter.sourceCode)
	if err != nil {
		panic(err.Error())
	}

	h.tree.Close()
	h.tree = tree
}

func HighlighterAdaptEditInput(event EventTextChange) (r sitter.EditInput) {
	// deletion
	if len(event.Text) == 0 {
		return sitter.EditInput{
			StartPoint:  sitter.Point{Row: uint32(event.Start.Line), Column: uint32(event.Start.Char)},
			OldEndPoint: sitter.Point{Row: uint32(event.End.Line), Column: uint32(event.End.Char)},
			NewEndPoint: sitter.Point{Row: uint32(event.Start.Line), Column: uint32(event.Start.Char)},
			StartIndex:  uint32(pointToByte(event.Buf, event.Start.Line, event.Start.Char)),
			OldEndIndex: uint32(pointToByte(event.Buf, event.Start.Line, event.Start.Char) + len(event.OldText)),
			NewEndIndex: uint32(pointToByte(event.Buf, event.Start.Line, event.Start.Char)),
		}
	}

	// insertion
	return sitter.EditInput{
		StartPoint:  sitter.Point{Row: uint32(event.Start.Line), Column: uint32(event.Start.Char)},
		OldEndPoint: sitter.Point{Row: uint32(event.Start.Line), Column: uint32(event.Start.Char)},
		NewEndPoint: sitter.Point{Row: uint32(event.NewEnd.Line), Column: uint32(event.NewEnd.Char)},
		StartIndex:  uint32(pointToByte(event.Buf, event.Start.Line, event.Start.Char)),
		OldEndIndex: uint32(pointToByte(event.Buf, event.Start.Line, event.Start.Char)),
		NewEndIndex: uint32(pointToByte(event.Buf, event.Start.Line, event.Start.Char) + utf8.RuneCountInString(event.Text)),
	}
}

func pointToByte(buf *Buffer, line, char int) int {
	size := 0
	lineNum := 0
	currentLine := buf.Lines.First()
	for currentLine != nil {
		if lineNum == line {
			v := currentLine.Value.Range(0, char)
			return size + utf8.RuneCountInString(string(v))
		}
		size += currentLine.Value.Bytes()
		currentLine = currentLine.Next()
		lineNum++
	}
	return size
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
	h.parser = sitter.NewParser()
	h.parser.SetLanguage(golang.GetLanguage())

	var err error

	hgFile := e.RuntimeDir("queries", "go", "highlights.scm")
	highlightQ, err := os.ReadFile(hgFile)
	if err != nil {
		fmt.Println(err)
	}

	h.q, err = sitter.NewQuery(highlightQ, golang.GetLanguage())
	if err != nil {
		h.e.LogError(errors.Wrap(err, "tree sitter query error"))
		return
	}

	h.Build()
	buf.Highlighter = h
}

func (h *Highlighter) Build() {
	tslock.Lock()
	defer tslock.Unlock()
	if h == nil {
		return
	}

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

// Get syntax highlights for document range
// TODO: this must return array of nodes
func (h *Highlighter) Highlights(lineStart, lineEnd uint32) {
	tslock.Lock()
	defer tslock.Unlock()

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

		// Apply predicates filteringsdkfsldfjslkdfj
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

