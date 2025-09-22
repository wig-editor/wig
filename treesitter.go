package wig

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/pkg/errors"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// TODO: rewrite treesitter to use channel and scheduled parsing. some day.
var tslock sync.Mutex

var _ Highlighter = &TreeSitterHighlighter{}

type TreeSitterHighlighter struct {
	e          *Editor
	buf        *Buffer
	parser     *sitter.Parser
	q          *sitter.Query
	tree       *sitter.Tree
	sourceCode []byte
}

func TreeSitterHighlighterGo(e *Editor) {
	go func() {
		for event := range e.Events.Subscribe() {
			switch e := event.Msg.(type) {
			case EventTextChange:
				if e.Buf.Highlighter != nil {
					e.Buf.Highlighter.TextChanged(e)
				}
			}
			event.Wg.Done()
		}
	}()
}

func (h *TreeSitterHighlighter) TextChanged(event EventTextChange) {
	if event.Buf == nil {
		return
	}

	tslock.Lock()
	defer tslock.Unlock()

	ll := h.editEditInput(event)
	h.tree.Edit(ll)

	h.sourceCode = []byte(event.Buf.String())
	tree, err := h.parser.ParseCtx(context.Background(), h.tree, h.sourceCode)
	if err != nil {
		// TODO: do not panic. log error.
		panic(err.Error())
	}

	h.tree.Close()
	h.tree = tree
}

func TreeSitterHighlighterInitBuffer(e *Editor, buf *Buffer) *TreeSitterHighlighter {
	if !strings.HasSuffix(buf.FilePath, ".go") {
		return nil
	}

	h := &TreeSitterHighlighter{
		e:   e,
		buf: buf,
	}
	h.parser = sitter.NewParser()
	h.parser.SetLanguage(golang.GetLanguage())

	var err error

	hgFile := e.RuntimeDir("queries", "go", "highlights.scm")
	highlightQ, err := os.ReadFile(hgFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	h.q, err = sitter.NewQuery(highlightQ, golang.GetLanguage())
	if err != nil {
		h.e.LogError(errors.Wrap(err, "tree sitter query error"))
		return nil
	}

	h.Build()
	return h
}

func (h *TreeSitterHighlighter) editEditInput(event EventTextChange) (r sitter.EditInput) {

	pointToByte := func(buf *Buffer, line, char int) int {
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

func (h *TreeSitterHighlighter) Build() {
	tslock.Lock()
	defer tslock.Unlock()
	if h == nil {
		return
	}

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

func (h *TreeSitterHighlighter) ForRange(lineStart, lineEnd uint32) *HighlighterCursor {
	tslock.Lock()
	defer tslock.Unlock()

	qc := sitter.NewQueryCursor()
	qc.SetPointRange(
		sitter.Point{Row: lineStart, Column: 0},
		sitter.Point{Row: lineEnd, Column: 0},
	)
	qc.Exec(h.q, h.tree.RootNode())
	defer qc.Close()

	nodes := List[HighlighterNode]{}
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		m = qc.FilterPredicates(m, h.sourceCode)
		for _, c := range m.Captures {
			startPoint := c.Node.StartPoint()
			endPoint := c.Node.EndPoint()
			nodes.PushBack(HighlighterNode{
				NodeName:  h.q.CaptureNameForId(c.Index),
				StartLine: startPoint.Row,
				StartChar: startPoint.Column,
				EndLine:   endPoint.Row,
				EndChar:   endPoint.Column,
			})
		}
	}

	return &HighlighterCursor{
		cursor: nodes.First(),
	}
}

