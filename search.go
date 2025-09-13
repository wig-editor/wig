package wig

import (
	"sort"

	str "github.com/boyter/go-string"
)

// Move cursor to the next search pattern match
func SearchNext(ctx Context, pattern string) {
	defer CmdEnsureCursorVisible(ctx)

	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)
	lineNum := cur.Line
	from := cur.Char + 1
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

		cur.Line = lineNum
		cur.Char = matches[0][0] + from
		cur.PreserveCharPosition = cur.Char
		break
	}
}

func SearchPrev(ctx Context, pattern string) {
	defer CmdEnsureCursorVisible(ctx)

	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)

	ln := cur.Line
	haystack := string(line.Value.Range(0, cur.Char-1))

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

		cur.Line = ln
		cur.Char = matches[0][0]
		cur.PreserveCharPosition = cur.Char
		break
	}
}

var LastSearchPattern string

func CmdSearchNext(ctx Context) {
	SearchNext(ctx, LastSearchPattern)
}

func CmdSearchPrev(ctx Context) {
	SearchPrev(ctx, LastSearchPattern)
}

