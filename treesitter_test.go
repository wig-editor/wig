package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
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
	TextInsert(buf, line, 22, " int")
	require.Equal(t, "func add(a int, b int) int {\n", line.Value.String())

	msg := <-events
	event := msg.(EventTextChange)
	require.Equal(t, EventTextChange{
		Buf:   buf,
		Start: Position{Line: 4, Char: 22},
		End:   Position{Line: 4, Char: 22},
		Text:  " int",
	}, event)
}

