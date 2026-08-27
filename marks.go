package wig

import (
	"strings"
	"unicode/utf8"
)

// Mark is a named cursor position bound to a specific buffer.
// When text is inserted or deleted in its buffer, the mark is adjusted
// so that it keeps pointing at the same text (vim-like behavior).
type Mark struct {
	Buf    *Buffer
	Cursor Cursor
}

// forEachWindowMark calls fn for every mark in every window that
// belongs to buf and stores the returned mark back into the window.
func forEachWindowMark(buf *Buffer, fn func(Mark) Mark) {
	for _, ws := range EditorInst.Workspaces {
		for _, win := range ws.Windows {
			if win == nil || win.Marks == nil {
				continue
			}
			for k, m := range win.Marks {
				if m.Buf == buf {
					win.Marks[k] = fn(m)
				}
			}
		}
	}
}

// adjustMarksForInsertion moves every mark belonging to ev.Buf past the
// inserted text, unless it is located before the insertion point. A mark
// located at or after the insertion point on the same line ends up after
// the inserted text (vim-like behavior).
//
// ev.Start is in pre-change buffer coordinates.
func adjustMarksForInsertion(ev EventTextChange) {
	if EditorInst == nil || ev.Buf == nil {
		return
	}

	text := ev.Text
	nRunes := utf8.RuneCountInString(text)
	nLines := strings.Count(text, "\n")
	lastLineRunes := utf8.RuneCountInString(text[strings.LastIndexByte(text, '\n')+1:])

	forEachWindowMark(ev.Buf, func(m Mark) Mark {
		cur := m.Cursor
		switch {
		case cur.Line < ev.Start.Line:
			// before the insertion: untouched
		case cur.Line == ev.Start.Line && cur.Char < ev.Start.Char:
			// before the insertion on the same line: untouched
		case cur.Line > ev.Start.Line:
			// below the insertion: shift down over the new lines
			cur.Line += nLines
		case nLines == 0:
			// same line, at or after the insertion point
			cur.Char += nRunes
		default:
			// same line as the insertion, at or after the insertion
			// point: everything after the point moves onto the last
			// line of the inserted block.
			cur.Line = ev.Start.Line + nLines
			cur.Char = cur.Char - ev.Start.Char + lastLineRunes
		}
		m.Cursor = cur
		return m
	})
}

// adjustMarksForDeletion moves every mark belonging to ev.Buf out of the
// deleted range [ev.Start, ev.End). Marks inside the range are collapsed
// to ev.Start, marks at or after ev.End are shifted back over the
// deleted text.
//
// ev.Start/ev.End are in pre-change buffer coordinates.
func adjustMarksForDeletion(ev EventTextChange) {
	if EditorInst == nil || ev.Buf == nil {
		return
	}

	s, e := ev.Start, ev.End
	deletedLines := e.Line - s.Line

	forEachWindowMark(ev.Buf, func(m Mark) Mark {
		cur := m.Cursor
		switch {
		case cur.Line < s.Line || (cur.Line == s.Line && cur.Char < s.Char):
			// before the deleted range: untouched
		case cur.Line < e.Line || (cur.Line == e.Line && cur.Char < e.Char):
			// inside the deleted range: collapse to its start
			cur.Line = s.Line
			cur.Char = s.Char
		case cur.Line == e.Line:
			// on the line where the deletion ends (at or after ev.End):
			// the surviving tail is spliced onto the line where the
			// deletion starts.
			cur.Line = s.Line
			cur.Char = s.Char + cur.Char - e.Char
		default:
			// after the deleted range: shift up over the removed lines
			cur.Line -= deletedLines
		}
		m.Cursor = cur
		return m
	})
}
