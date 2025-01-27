package mcwig

import (
	"testing"

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
