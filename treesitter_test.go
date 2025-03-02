package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/require"
)

func TestTreeSitterNodeCursor(t *testing.T) {
	nodes := List[TreeSitterRangeNode]{}

	nodes.PushBack(TreeSitterRangeNode{
		NodeName:  "test0",
		StartLine: 0,
		StartChar: 0,
		EndLine:   0,
		EndChar:   4,
	})

	nodes.PushBack(TreeSitterRangeNode{
		NodeName:  "test1",
		StartLine: 0,
		StartChar: 6,
		EndLine:   0,
		EndChar:   10,
	})

	nodes.PushBack(TreeSitterRangeNode{
		NodeName:  "test2",
		StartLine: 1,
		StartChar: 2,
		EndLine:   1,
		EndChar:   5,
	})

	cur := NewColorNodeCursor(nodes.First())

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
	source := `package mcwig

import "fmt"

func add(a int, b int) {
	fmt.Printf("%d", a+b)
}`

	e := NewEditor(
		testutils.Viewport,
		nil,
	)
	buf := e.BufferFindByFilePath("testfile", true)
	buf.ResetLines()
	buf.Append(source)
	require.Equal(t, source+"\n", buf.String())

	events := e.Events.Subscribe()

	line := CursorLineByNum(buf, 4)
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 4, Char: 5},
		End:   Cursor{Line: 4, Char: 8},
	})
	require.Equal(t, "func (a int, b int) {\n", line.Value.String())

	msg := <-events
	event := msg.(EventTextChange)
	require.Equal(t, EventTextChange{
		Buf:     buf,
		Start:   Position{Line: 4, Char: 5},
		End:     Position{Line: 4, Char: 8},
		Text:    "",
		OldText: "add",
	}, event)

	expected := sitter.EditInput{
		StartPoint:  sitter.Point{Row: 4, Column: 5},
		OldEndPoint: sitter.Point{Row: 4, Column: 8},
		NewEndPoint: sitter.Point{Row: 4, Column: 5},
		StartIndex:  uint32(34),
		OldEndIndex: uint32(37),
		NewEndIndex: uint32(34),
	}

	actual := HighlighterAdaptEditInput(event)
	require.Equal(t, expected, actual)
}

func TestTreeSitter_AdaptEventTextChangeDeleteLine(t *testing.T) {
	source := `package mcwig

import "fmt"

func add(a int, b int) {
	fmt.Printf("%d", a+b)
}`

	e := NewEditor(
		testutils.Viewport,
		nil,
	)
	buf := e.BufferFindByFilePath("testfile", true)
	buf.ResetLines()
	buf.Append(source)
	require.Equal(t, source+"\n", buf.String())

	events := e.Events.Subscribe()

	buf.Cursor.Line = 4
	buf.Cursor.Char = 0

	CmdDeleteLine(Context{
		Editor: e,
		Buf:    buf,
		Count:  0,
		Char:   "",
	})

	msg := <-events
	event := msg.(EventTextChange)
	require.Equal(t, EventTextChange{
		Buf:     buf,
		Start:   Position{Line: 4, Char: 0},
		End:     Position{Line: 5, Char: 0},
		Text:    "",
		OldText: "func add(a int, b int) {\n",
	}, event)

	expected := sitter.EditInput{
		StartPoint:  sitter.Point{Row: 4, Column: 0},
		OldEndPoint: sitter.Point{Row: 5, Column: 0},
		NewEndPoint: sitter.Point{Row: 4, Column: 0},
		StartIndex:  uint32(29),
		OldEndIndex: uint32(54),
		NewEndIndex: uint32(29),
	}

	actual := HighlighterAdaptEditInput(event)
	require.Equal(t, expected, actual)
}

func TestTreeSitter_AdaptEventTextInsert(t *testing.T) {
	source := `package mcwig

import "fmt"

func add(a int, b int) {
	fmt.Printf("%d", a+b)
}`

	e := NewEditor(
		testutils.Viewport,
		nil,
	)
	buf := e.BufferFindByFilePath("testfile", true)
	buf.ResetLines()
	buf.Append(source)
	require.Equal(t, source+"\n", buf.String())

	events := e.Events.Subscribe()

	buf.Cursor.Line = 4
	buf.Cursor.Char = 0

	msg := <-events
	event := msg.(EventTextChange)
	require.Equal(t, EventTextChange{
		Buf:     buf,
		Start:   Position{Line: 4, Char: 8},
		End:     Position{Line: 4, Char: 8},
		Text:    "1",
		OldText: "",
	}, event)

	expected := sitter.EditInput{
		StartPoint:  sitter.Point{Row: 4, Column: 8},
		OldEndPoint: sitter.Point{Row: 5, Column: 8},
		NewEndPoint: sitter.Point{Row: 4, Column: 9},
		StartIndex:  uint32(37),
		OldEndIndex: uint32(37),
		NewEndIndex: uint32(38),
	}

	actual := HighlighterAdaptEditInput(event)
	require.Equal(t, expected, actual)
}

