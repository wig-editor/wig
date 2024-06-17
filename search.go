package mcwig

import (
	"sort"

	str "github.com/boyter/go-string"
)

// Move cursor to the next search patten match
func SearchNext(e *Editor, buf *Buffer, line *Element[Line], pattern string) {
	defer CmdEnsureCursorVisible(e)

	lineNum := buf.Cursor.Line
	from := buf.Cursor.Char + 1
	haystack := string(line.Value.Range(from, EOL))

	for line != nil {
		matches := str.IndexAllIgnoreCase(haystack, pattern, 1)
		if len(matches) == 0 {
			line = line.Next()
			if line == nil {
				break
			}
			lineNum++
			from = 0
			haystack = string(line.Value)
			continue
		}

		sort.Slice(matches, func(i, j int) bool {
			return matches[i][0] < matches[j][0]
		})

		buf.Cursor.Line = lineNum
		buf.Cursor.Char = matches[0][0] + from
		break
	}
}

func SearchPrev(e *Editor, buf *Buffer, line *Element[Line], pattern string) {
	defer CmdEnsureCursorVisible(e)

	ln := buf.Cursor.Line
	haystack := string(line.Value.Range(0, buf.Cursor.Char-1))

	for line != nil {
		matches := str.IndexAllIgnoreCase(haystack, pattern, -1)
		if len(matches) == 0 {
			line = line.Prev()
			if line == nil {
				break
			}
			ln--
			haystack = string(line.Value)
			continue
		}

		sort.Slice(matches, func(i, j int) bool {
			return matches[i][0] > matches[j][0]
		})

		buf.Cursor.Line = ln
		buf.Cursor.Char = matches[0][0]
		break
	}
}

var LastSearchPattern string

func CmdSearchNext(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SearchNext(e, buf, line, LastSearchPattern)
	})
}

func CmdSearchPrev(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		SearchPrev(e, buf, line, LastSearchPattern)
	})
}
