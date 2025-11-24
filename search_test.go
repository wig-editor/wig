package wig

import (
	"testing"

	"github.com/firstrow/wig/testutils"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf, _ := e.OpenFile(testutils.Filepath("buffer_test.txt"))

	ctx := Context{
		Editor: e,
		Buf:    buf,
	}
	cur := CursorGet(e, buf)

	SearchNext(ctx, "one")
	require.Equal(t, 0, cur.Line)
	require.Equal(t, 5, cur.Char)

	SearchNext(ctx, "three")
	require.Equal(t, 2, cur.Line)
	require.Equal(t, 5, cur.Char)

	SearchPrev(ctx, "one")
	require.Equal(t, 0, cur.Line)
	require.Equal(t, 5, cur.Char)
}

