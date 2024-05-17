package mcwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	buf := NewBuffer()
	buf.Lines.Insert([]rune("hello"), 0)

	assert.Equal(t, 1, buf.Lines.Size)
}

func TestBufferReadFile(t *testing.T) {
	buf, err := BufferReadFile("/home/andrew/code/mcwig/buffer_test.txt")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	assert.Equal(t, 6, buf.Lines.Size)
}

func TestLineByNum(t *testing.T) {
	buf, err := BufferReadFile("/home/andrew/code/mcwig/buffer_test.txt")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	line := lineByNum(buf, 1)

	assert.Equal(t, "line two", string(line.Data))
}

func TestSelectionDelete(t *testing.T) {
	// buf, err := BufferReadFile("/home/andrew/code/mcwig/buffer_test.txt")
	// if err != nil {
	// 	t.Errorf("expected nil, got %v", err)
	// }
}
