package mcwig

import (
	"strings"
	"unicode"
)

func indent(ctx Context) {
	if !strings.HasSuffix(ctx.Buf.FilePath, ".go") {
		return
	}

	line := CursorLine(ctx.Buf)

	prevLine := line.Prev()
	for prevLine != nil {
		if prevLine.Value.IsEmpty() {
			prevLine = prevLine.Prev()
			continue
		}

		idx := 0
		for i, c := range prevLine.Value {
			if !unicode.IsSpace(c) {
				idx = i
				break
			}
		}

		trimmed := strings.TrimSpace(string(prevLine.Value))
		if strings.HasSuffix(trimmed, "{") {
			idx += 1
		}

		ch := strings.Repeat("\t", idx)
		TextInsert(ctx.Buf, line, 0, ch)
		CmdGotoLineEnd(ctx)

		break
	}
}

// Get number if "indents" in provided line
// Indent unit can be \t, or any number of spaces. eg. 2 or 4.
func IndentGetNumber(line []rune, indentUnit []rune) int {
	fullStep := len(indentUnit)
	if fullStep == 0 || len(line) == 0 {
		return 0
	}

	unit := indentUnit[0]
	i := 0
	count := 0

	for _, ch := range line {
		if ch != unit {
			break
		}

		i++

		if i == fullStep {
			i = 0
			count++
		}
	}

	return count
}

