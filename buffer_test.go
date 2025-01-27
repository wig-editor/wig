package mcwig

import (
	"os"
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 6, buf.Lines.Len)
}

func TestLineByNum(t *testing.T) {
	buf, err := BufferReadFile("/home/andrew/code/mcwig/buffer_test.txt")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	line := CursorLineByNum(buf, 1)
	assert.Equal(t, "line two", string(line.Value))
}

func TestSelectionDelete(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	buf.Selection = &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 0},
	}
	CmdSelectinDelete(e)
	line := CursorLineByNum(buf, 0)
	assert.Equal(t, "ine two", string(line.Value))
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
	assert.Equal(t, 6, buf.Lines.Len)
}

func TestWordUnderCusor(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	e.ActiveWindow().ShowBuffer(buf)
	buf.Selection = &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 0},
	}
	CmdSelectinDelete(e)
	line := CursorLineByNum(buf, 0)
	assert.Equal(t, "ine two", string(line.Value))
}

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
