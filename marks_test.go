package wig

import (
	"testing"

	"github.com/firstrow/wig/testutils"
	"github.com/stretchr/testify/require"
)

// markTestEnv creates an editor with a buffer filled with the given
// content and returns the buffer together with the active window.
func markTestEnv(t *testing.T, content string) (*Buffer, *Window) {
	t.Helper()
	e := NewEditor(testutils.Viewport, nil)
	buf := e.BufferFindByFilePath("mark-test", true)
	buf.ResetLines()
	buf.Append(content)
	require.Equal(t, content, buf.String())
	return buf, e.ActiveWindow()
}

func setMark(win *Window, buf *Buffer, r rune, line, char int) {
	win.Marks[r] = Mark{Buf: buf, Cursor: Cursor{Line: line, Char: char}}
}

func TestMarkMovesOnCharInsertion(t *testing.T) {
	tests := []struct {
		name     string
		insertAt Position
		text     string
		markAt   Cursor
		expected Cursor
	}{
		{
			name:     "insertion before mark on same line",
			insertAt: Position{Line: 0, Char: 1},
			text:     "XY",
			markAt:   Cursor{Line: 0, Char: 4},
			expected: Cursor{Line: 0, Char: 6},
		},
		{
			name:     "insertion at mark position moves mark past inserted text",
			insertAt: Position{Line: 0, Char: 2},
			text:     "X",
			markAt:   Cursor{Line: 0, Char: 2},
			expected: Cursor{Line: 0, Char: 3},
		},
		{
			name:     "insertion after mark on same line",
			insertAt: Position{Line: 0, Char: 3},
			text:     "XY",
			markAt:   Cursor{Line: 0, Char: 1},
			expected: Cursor{Line: 0, Char: 1},
		},
		{
			name:     "insertion on line above",
			insertAt: Position{Line: 0, Char: 0},
			text:     "top",
			markAt:   Cursor{Line: 1, Char: 2},
			expected: Cursor{Line: 1, Char: 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf, win := markTestEnv(t, "hello\nworld\n")
			setMark(win, buf, 'a', tc.markAt.Line, tc.markAt.Char)
			line := CursorLineByNum(buf, tc.insertAt.Line)
			TextInsert(buf, line, tc.insertAt.Char, tc.text)
			require.Equal(t, tc.expected, win.Marks['a'].Cursor)
		})
	}
}

func TestMarkMovesOnLineInsertion(t *testing.T) {
	buf, win := markTestEnv(t, "one\ntwo\nthree\n")
	setMark(win, buf, 'a', 2, 1)

	// insert a new line between "one" and "two"
	TextInsert(buf, CursorLineByNum(buf, 0), 3, "\n")
	require.Equal(t, "one\n\ntwo\nthree\n", buf.String())
	require.Equal(t, Cursor{Line: 3, Char: 1}, win.Marks['a'].Cursor)
}

func TestMarkMovesOnMultilineInsertion(t *testing.T) {
	buf, win := markTestEnv(t, "hello\nworld\n")

	// "hello" -> "helAB" / "CDlo", "world"
	setMark(win, buf, 'a', 0, 4)
	TextInsert(buf, CursorLineByNum(buf, 0), 3, "AB\nCD")
	require.Equal(t, "helAB\nCDlo\nworld\n", buf.String())
	require.Equal(t, Cursor{Line: 1, Char: 3}, win.Marks['a'].Cursor)

	// mark at the insertion point ends up after the inserted text
	setMark(win, buf, 'b', 1, 2)
	TextInsert(buf, CursorLineByNum(buf, 1), 2, "X")
	require.Equal(t, Cursor{Line: 1, Char: 3}, win.Marks['b'].Cursor)
}

func TestMarkMovesOnCharDeletion(t *testing.T) {
	tests := []struct {
		name     string
		sel      Selection
		markAt   Cursor
		expected Cursor
	}{
		{
			name:     "deletion before mark on same line",
			sel:      Selection{Start: Cursor{Line: 0, Char: 1}, End: Cursor{Line: 0, Char: 3}},
			markAt:   Cursor{Line: 0, Char: 4},
			expected: Cursor{Line: 0, Char: 2},
		},
		{
			name:     "deletion at mark position collapses mark to deletion start",
			sel:      Selection{Start: Cursor{Line: 0, Char: 2}, End: Cursor{Line: 0, Char: 4}},
			markAt:   Cursor{Line: 0, Char: 2},
			expected: Cursor{Line: 0, Char: 2},
		},
		{
			name:     "mark at end of deleted range moves to deletion start",
			sel:      Selection{Start: Cursor{Line: 0, Char: 0}, End: Cursor{Line: 0, Char: 4}},
			markAt:   Cursor{Line: 0, Char: 4},
			expected: Cursor{Line: 0, Char: 0},
		},
		{
			name:     "deletion after mark on same line",
			sel:      Selection{Start: Cursor{Line: 0, Char: 3}, End: Cursor{Line: 0, Char: 5}},
			markAt:   Cursor{Line: 0, Char: 1},
			expected: Cursor{Line: 0, Char: 1},
		},
		{
			name:     "deletion on other line",
			sel:      Selection{Start: Cursor{Line: 1, Char: 0}, End: Cursor{Line: 1, Char: 2}},
			markAt:   Cursor{Line: 0, Char: 3},
			expected: Cursor{Line: 0, Char: 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf, win := markTestEnv(t, "hello\nworld\n")
			setMark(win, buf, 'a', tc.markAt.Line, tc.markAt.Char)
			TextDelete(buf, &tc.sel)
			require.Equal(t, tc.expected, win.Marks['a'].Cursor)
		})
	}
}

func TestMarkMovesOnLineDeletion(t *testing.T) {
	buf, win := markTestEnv(t, "line one\nline two\nline three\n")
	setMark(win, buf, 'a', 2, 5)

	// delete the whole first line (including its trailing "\n")
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 0, Char: 9},
	})
	require.Equal(t, "line two\nline three\n", buf.String())
	require.Equal(t, Cursor{Line: 1, Char: 5}, win.Marks['a'].Cursor)
}

func TestMarkMovesOnMultilineDeletion(t *testing.T) {
	buf, win := markTestEnv(t, "abcdef\nghij\nklm\n")
	setMark(win, buf, 'a', 2, 2)
	setMark(win, buf, 'b', 1, 3)
	setMark(win, buf, 'c', 1, 1)

	// delete "abcdef\n" and "gh": lines get spliced
	TextDelete(buf, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 2},
	})
	require.Equal(t, "ij\nklm\n", buf.String())

	// 'b' pointed at 'j' in "ghij" -> now second char of "ij"
	require.Equal(t, Cursor{Line: 0, Char: 1}, win.Marks['b'].Cursor)
	// 'c' pointed at 'h' inside the deleted range -> collapsed to start
	require.Equal(t, Cursor{Line: 0, Char: 0}, win.Marks['c'].Cursor)
	// 'a' pointed at 'm' in "klm" -> shifted up one line
	require.Equal(t, Cursor{Line: 1, Char: 2}, win.Marks['a'].Cursor)
}

func TestMarkUnaffectedByOtherBufferChanges(t *testing.T) {
	buf, win := markTestEnv(t, "hello\nworld\n")
	other := EditorInst.BufferFindByFilePath("mark-test-other", true)
	other.ResetLines()
	other.Append("aaa\nbbb\n")

	setMark(win, buf, 'a', 1, 2)

	TextInsert(other, other.Lines.First(), 0, "zz")
	TextDelete(other, &Selection{
		Start: Cursor{Line: 0, Char: 0},
		End:   Cursor{Line: 1, Char: 0},
	})
	require.Equal(t, Cursor{Line: 1, Char: 2}, win.Marks['a'].Cursor)
}
