package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/assert"
)

func TestYankSingleLine(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	buf := e.ActiveBuffer()
	buf.Selection = nil
	CmdYank(e)
	CmdYank(e)
	CmdYank(e)

	assert.Equal(t, 1, buf.Yanks.Len)

	CmdYankPut(e)

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
	e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	buf := e.ActiveBuffer()
	// "ine thre"
	buf.Selection = &Selection{Start: Cursor{Line: 2, Char: 1}, End: Cursor{Line: 2, Char: 8}}
	CmdYank(e)
	buf.Cursor = Cursor{Line: 0, Char: 3}

	assert.Equal(t, 1, buf.Yanks.Len)

	CmdYankPut(e)

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
	CmdYank(e)
	CmdYankPut(e)
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
	CmdYank(e)
	buf.Cursor = Cursor{Line: 5, Char: 3}
	CmdYankPutBefore(e)
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
