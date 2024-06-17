package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	buf := e.Buffers[0]

	SearchNext(e, buf, buf.Lines.First(), "one")
	require.Equal(t, 0, buf.Cursor.Line)
	require.Equal(t, 5, buf.Cursor.Char)

	SearchNext(e, buf, buf.Lines.First(), "three")
	require.Equal(t, 2, buf.Cursor.Line)
	require.Equal(t, 5, buf.Cursor.Char)

	SearchPrev(e, buf, CursorLineByNum(buf, 2), "one")
	require.Equal(t, 0, buf.Cursor.Line)
	require.Equal(t, 5, buf.Cursor.Char)
}
