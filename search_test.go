package wig

import (
	"testing"

	"github.com/firstrow/wig/testutils"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile(testutils.Filepath("buffer_test.txt"))

	ctx := Context{
		Editor: e,
		Buf:    buf,
	}

	SearchNext(ctx, "one")
	require.Equal(t, 0, buf.Cursor.Line)
	require.Equal(t, 5, buf.Cursor.Char)

	SearchNext(ctx, "three")
	require.Equal(t, 2, buf.Cursor.Line)
	require.Equal(t, 5, buf.Cursor.Char)

	SearchPrev(ctx, "one")
	require.Equal(t, 0, buf.Cursor.Line)
	require.Equal(t, 5, buf.Cursor.Char)
}

