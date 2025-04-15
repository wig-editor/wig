package mcwig

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/firstrow/mcwig/testutils"
)

func TestYankSingleLine(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile(testutils.Filepath("buffer_test.txt"))
	ctx := Context{
		Editor: e,
		Buf:    buf,
	}

	e.ActiveWindow().ShowBuffer(buf)
	buf.Selection = nil

	CmdYank(ctx)
	CmdYank(ctx)
	CmdYank(ctx)

	assert.Equal(t, 1, e.Yanks.Len)

	CmdYankPut(ctx)

	expected := `line one
line one
line two
line three
line four
line five
`
	assert.Equal(t, expected, buf.String())
}

func TestYankSelection(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile(testutils.Filepath("buffer_test.txt"))
	ctx := Context{
		Editor: e,
		Buf:    buf,
	}
	e.ActiveWindow().ShowBuffer(buf)
	// "ine thre"
	buf.Selection = &Selection{Start: Cursor{Line: 2, Char: 1}, End: Cursor{Line: 2, Char: 8}}
	CmdYank(ctx)
	buf.Cursor = Cursor{Line: 0, Char: 3}

	assert.Equal(t, 1, e.Yanks.Len)

	CmdYankPut(ctx)

	expected := `lineine thre one
line two
line three
line four
line five
`
	assert.Equal(t, expected, buf.String())

	// test line paste
	buf.Cursor = Cursor{Line: 2, Char: 3}
	buf.Selection = nil
	CmdYank(ctx)
	CmdYankPut(ctx)
	expected = `lineine thre one
line two
line three
line three
line four
line five
`
	assert.Equal(t, expected, buf.String())

	// put above
	buf.Cursor = Cursor{Line: 1, Char: 3}
	buf.Selection = nil
	CmdYank(ctx)
	buf.Cursor = Cursor{Line: 5, Char: 3}
	CmdYankPutBefore(ctx)
	expected = `lineine thre one
line two
line three
line three
line four
line two
line five
`
	assert.Equal(t, expected, buf.String())
}

