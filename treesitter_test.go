package wig

import (
	"sync"
	"testing"

	"github.com/firstrow/wig/testutils"
	"github.com/stretchr/testify/require"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

// TODO: rewrite/fix treesitter concurrency tests

func TestTreeSitterNodeCursor(t *testing.T) {
	nodes := List[HighlighterNode]{}

	nodes.PushBack(HighlighterNode{
		NodeName:  "test0",
		StartLine: 0,
		StartChar: 0,
		EndLine:   0,
		EndChar:   4,
	})

	nodes.PushBack(HighlighterNode{
		NodeName:  "test1",
		StartLine: 0,
		StartChar: 6,
		EndLine:   0,
		EndChar:   10,
	})

	nodes.PushBack(HighlighterNode{
		NodeName:  "test2",
		StartLine: 1,
		StartChar: 2,
		EndLine:   1,
		EndChar:   5,
	})

	cur := HighlighterCursor{nodes.First()}

	node, ok := cur.Seek(0, 0)
	require.Equal(t, true, ok)
	require.Equal(t, "test0", node.Value.NodeName)

	node, ok = cur.Seek(0, 3)
	require.Equal(t, true, ok)
	require.Equal(t, "test0", node.Value.NodeName)

	_, ok = cur.Seek(0, 5)
	require.Equal(t, false, ok)

	node, ok = cur.Seek(0, 6)
	require.Equal(t, true, ok)
	require.Equal(t, "test1", node.Value.NodeName)

	node, ok = cur.Seek(0, 9)
	require.Equal(t, true, ok)
	require.Equal(t, "test1", node.Value.NodeName)

	_, ok = cur.Seek(1, 1)
	require.Equal(t, false, ok)

	node, ok = cur.Seek(1, 3)
	require.Equal(t, true, ok)
	require.Equal(t, "test2", node.Value.NodeName)
}

func TestTreeSitter_AdaptEventTextChange(t *testing.T) {
	source := `package wig

import "fmt"

func add(a int, b int) {
	fmt.Printf("%d", a+b)
}`

	e := NewEditor(
		testutils.Viewport,
		nil,
	)
	buf := e.BufferFindByFilePath("testfile.go", true)
	buf.ResetLines()
	buf.Append(source)
	require.Equal(t, source+"\n", buf.String())
	buf.Highlighter = TreeSitterHighlighterInitBuffer(e, buf)
	highlighter := buf.Highlighter.(*TreeSitterHighlighter)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		events := e.Events.Subscribe()
		wg.Done()
		msg := <-events
		msg.Wg.Done()
		event := msg.Msg.(EventTextChange)
		require.Equal(t, EventTextChange{
			Buf:     buf,
			Start:   Position{Line: 4, Char: 5},
			End:     Position{Line: 4, Char: 8},
			Text:    "",
			OldText: "add",
		}, event)

		actual := highlighter.editEditInput(event)
		expected := sitter.InputEdit{
			StartPosition:  sitter.Point{Row: 4, Column: 5},
			OldEndPosition: sitter.Point{Row: 4, Column: 8},
			NewEndPosition: sitter.Point{Row: 4, Column: 5},
			StartByte:      uint(34),
			OldEndByte:     uint(37),
			NewEndByte:     uint(34),
		}
		require.Equal(t, expected, actual)
	}()
	wg.Wait()

	line := CursorLineByNum(buf, 4)
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 4, Char: 5},
		End:   Cursor{Line: 4, Char: 8},
	})
	require.Equal(t, "func (a int, b int) {\n", line.Value.String())

}

func TestTreeSitter_AdaptEventTextChangeDeleteLine(t *testing.T) {
	source := `package wig

import "fmt"

func add(a int, b int) {
	fmt.Printf("%d", a+b)
}`

	e := NewEditor(
		testutils.Viewport,
		nil,
	)

	buf := e.BufferFindByFilePath("testfile.go", true)
	buf.ResetLines()
	buf.Append(source)
	cur := CursorGet(e, buf)
	cur.Line = 4
	cur.Char = 0
	require.Equal(t, source+"\n", buf.String())
	buf.Highlighter = TreeSitterHighlighterInitBuffer(e, buf)
	highlighter := buf.Highlighter.(*TreeSitterHighlighter)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		events := e.Events.Subscribe()
		wg.Done()
		msg := <-events
		msg.Wg.Done()
		msg = <-events
		msg.Wg.Done()
		event := msg.Msg.(EventTextChange)
		require.Equal(t, EventTextChange{
			Buf:     buf,
			Start:   Position{Line: 4, Char: 0},
			End:     Position{Line: 5, Char: 0},
			Text:    "",
			OldText: "func add(a int, b int) {\n",
		}, event)

		expected := sitter.InputEdit{
			StartPosition:  sitter.Point{Row: 4, Column: 0},
			OldEndPosition: sitter.Point{Row: 5, Column: 0},
			NewEndPosition: sitter.Point{Row: 4, Column: 0},
			StartByte:      uint(29),
			OldEndByte:     uint(54),
			NewEndByte:     uint(29),
		}

		actual := highlighter.editEditInput(event)
		require.Equal(t, expected, actual)
	}()
	wg.Wait()

	CmdDeleteLine(Context{
		Editor: e,
		Buf:    buf,
		Count:  0,
		Char:   "",
	})
}

