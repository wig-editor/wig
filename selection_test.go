package mcwig

import (
	"testing"
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
