package wig

import (
	"strings"
	"testing"

	"github.com/firstrow/wig/testutils"
	"github.com/stretchr/testify/require"
)

func TestTextObjects(t *testing.T) {
	return

	e := NewEditor(
		testutils.Viewport,
		nil,
	)
	buf := NewBuffer()
	input := "( ok )"
	buf.Append(input)
	buf.Cursor.Line = 0
	buf.Cursor.Char = 2

	e.Buffers = append(e.Buffers, buf)

	type test struct {
		cursor   Cursor
		lines    string
		ch       rune
		include  bool
		found    bool
		expected *Selection
	}

	tcases := []test{
		{
			cursor:  Cursor{Line: 0, Char: 3},
			lines:   "(ok)",
			ch:      '(',
			include: true,
			found:   true,
			expected: &Selection{
				Start: Cursor{0, 0, 0},
				End:   Cursor{0, 3, 0},
			},
		},
		{
			cursor:   Cursor{Line: 1, Char: 0},
			lines:    "ok)\n)",
			ch:       '(',
			include:  true,
			found:    false,
			expected: nil,
		},
		{
			cursor:  Cursor{Line: 0, Char: 0},
			lines:   "()",
			ch:      '(',
			include: true,
			found:   true,
			expected: &Selection{
				Start: Cursor{0, 0, 0},
				End:   Cursor{0, 1, 0},
			},
		},
		{
			cursor:   Cursor{Line: 0, Char: 0},
			lines:    "()",
			ch:       '(',
			include:  false,
			found:    true,
			expected: nil,
		},
		{
			cursor:  Cursor{Line: 0, Char: 6},
			lines:   "( (  ) )",
			ch:      '(',
			include: false,
			found:   true,
			expected: &Selection{
				Start: Cursor{0, 1, 0},
				End:   Cursor{0, 6, 0},
			},
		},
	}

	for _, c := range tcases {
		buf.ResetLines()
		lines := strings.Split(c.lines, "\n")
		for _, line := range lines {
			buf.Append(line)
		}
		require.Equal(t, c.lines, buf.String())
		buf.Cursor = c.cursor

		found, sel, _ := TextObjectBlock(buf, c.ch, c.include)
		require.Equal(t, c.found, found, c)
		require.Equal(t, c.cursor.Char, buf.Cursor.Char)
		require.Equal(t, c.cursor.Line, buf.Cursor.Line)

		if found && c.expected != nil {
			require.Equal(t, c.expected.Start.Line, sel.Start.Line, c)
			require.Equal(t, c.expected.Start.Char, sel.Start.Char, c)
			require.Equal(t, c.expected.End.Line, sel.End.Line, c)
			require.Equal(t, c.expected.End.Char, sel.End.Char, c)
		}
	}

}
