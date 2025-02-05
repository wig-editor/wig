package mcwig

import (
	"os"
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuffer(t *testing.T) {
	buf := NewBuffer()
	assert.Equal(t, 1, buf.Lines.Len)
	buf.Lines.PushFront(Line{})
	assert.Equal(t, 2, buf.Lines.Len)
}

func TestBufferReadFile(t *testing.T) {
	buf, err := BufferReadFile("/home/andrew/code/mcwig/buffer_test.txt")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	assert.Equal(t, 5, buf.Lines.Len)
}

func TestLineByNum(t *testing.T) {
	buf, err := BufferReadFile("/home/andrew/code/mcwig/buffer_test.txt")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	line := CursorLineByNum(buf, 1)
	assert.Equal(t, "line two\n", string(line.Value))
}

func TestSelectionDelete(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	buf.Selection = &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 0},
	}
	ctx := Context{
		Editor: e,
		Buf:    buf,
	}
	SelectionDelete(ctx)
	line := CursorLineByNum(buf, 0)
	assert.Equal(t, "ine two\n", string(line.Value))
}

func TestSaveFile(t *testing.T) {
	tmpFilePath := "/tmp/wcwig_test.go"
	testFilePath := "/home/andrew/code/mcwig/buffer_test.txt"

	err := copyFile(testFilePath, tmpFilePath)
	assert.NoError(t, err)

	buf, err := BufferReadFile(tmpFilePath)
	assert.NoError(t, err)

	err = buf.Save()
	err = buf.Save()
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	buf, err = BufferReadFile(tmpFilePath)
	assert.NoError(t, err)
	assert.Equal(t, 5, buf.Lines.Len)
}

func TestWordUnderCusor(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	buf.Selection = &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 0},
	}
	ctx := Context{
		Editor: e,
		Buf:    buf,
	}
	SelectionDelete(ctx)
	line := CursorLineByNum(buf, 0)
	assert.Equal(t, "ine two\n", string(line.Value))
}

func TestTextInsertNewLine(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())

	expected := `line
 one
line two
line three
line four
line five
`

	TextInsert(buf, buf.Lines.First(), 4, "\n")
	require.Equal(t, expected, string(buf.String()))
}

func TestTextInsertDelete(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())

	TextInsert(buf, buf.Lines.First(), 4, " test")
	require.Equal(t, "line test one\n", string(buf.Lines.First().Value))

	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 4},
		End:   Cursor{Line: 0, Char: 9},
	})
	require.Equal(t, "line one\n", string(buf.Lines.First().Value))
}

func TestTextDeleteMultiline(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 0},
	})
	expected := `line two
line three
line four
line five
`
	require.Equal(t, expected, buf.String())
}

func TestTextDeleteMultiline2(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 2, Char: 5},
	})
	expected := `three
line four
line five
`
	require.Equal(t, expected, buf.String())
}

func TestTextDeleteToEOL(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 0, Char: 8},
	})
	expected := `
line two
line three
line four
line five
`
	require.Equal(t, expected, buf.String())
}

func TestTextDelete_EOL(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 0, Char: 9},
	})
	expected := `line two
line three
line four
line five
`
	require.Equal(t, expected, buf.String())
}

func TestTextDelete_EOL_EOF(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	defer CmdKillBuffer(e.NewContext())
	// delete last line
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 4, Char: 0},
		End:   Cursor{Line: 4, Char: 10},
	})
	expected := `line one
line two
line three
line four
`
	require.Equal(t, expected, buf.String())
}

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
