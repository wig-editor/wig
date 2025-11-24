package wig

import (
	"strings"
	"unicode"
)

func indent(ctx Context) {
	indentChars := []string{"{", ":"}
	// outdentChars := []string{"}"}

	cur := ContextCursorGet(ctx)
	line := CursorLine(ctx.Buf, cur)

	prevLine := line.Prev()
	for prevLine != nil {
		if prevLine.Value.IsEmpty() {
			prevLine = prevLine.Prev()
			continue
		}

		idx := 0
		indentCh := "\t"
		for i, c := range prevLine.Value {
			if !unicode.IsSpace(c) {
				idx = i
				break
			}
			indentCh = string(c)
		}

		trimmed := strings.TrimSpace(string(prevLine.Value))
		for _, ch := range indentChars {
			if strings.HasSuffix(trimmed, ch) {
				idx += 1
			}
		}

		ch := strings.Repeat(indentCh, idx)
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
