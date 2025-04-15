package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSelectionCursorInRange(t *testing.T) {
	tests := []struct {
		sel  Selection
		c    Cursor
		want bool
	}{
		{sel: Selection{Start: Cursor{Line: 1, Char: 10}, End: Cursor{Line: 2, Char: 20}}, c: Cursor{Line: 1, Char: 10}, want: true},
		{sel: Selection{Start: Cursor{Line: 1, Char: 10}, End: Cursor{Line: 2, Char: 20}}, c: Cursor{Line: 2, Char: 20}, want: true},
		{sel: Selection{Start: Cursor{Line: 1, Char: 10}, End: Cursor{Line: 2, Char: 20}}, c: Cursor{Line: 0, Char: 0}, want: false},
		{sel: Selection{Start: Cursor{Line: 1, Char: 10}, End: Cursor{Line: 2, Char: 20}}, c: Cursor{Line: 3, Char: 30}, want: false},
		{sel: Selection{Start: Cursor{Line: 1, Char: 10}, End: Cursor{Line: 2, Char: 20}}, c: Cursor{Line: 1, Char: 25}, want: true},
		{sel: Selection{Start: Cursor{Line: 1, Char: 10}, End: Cursor{Line: 2, Char: 20}}, c: Cursor{Line: 2, Char: 15}, want: true},
	}

	for _, tt := range tests {
		if got := SelectionCursorInRange(&tt.sel, tt.c); got != tt.want {
			t.Errorf("SelectionCursorInRange(%v, %v) = %t, want %t", tt.sel, tt.c, got, tt.want)
		}
	}
}

func TestSelectionToString(t *testing.T) {
	buf, err := BufferReadFile(testutils.Filepath("buffer_test.txt"))
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	buf.Selection = &Selection{Start: Cursor{Line: 2, Char: 1}, End: Cursor{Line: 2, Char: 8}}
	assert.Equal(t, "ine thre", SelectionToString(buf, buf.Selection))

	buf.Selection = &Selection{Start: Cursor{Line: 0, Char: 0}, End: Cursor{Line: 0, Char: 1111}}
	assert.Equal(t, "line one\n", SelectionToString(buf, buf.Selection))

	buf.Selection = &Selection{Start: Cursor{Line: 1, Char: 0}, End: Cursor{Line: 2, Char: 3}}
	assert.Equal(t, "line two\nline", SelectionToString(buf, buf.Selection))
}

