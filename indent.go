package mcwig

import "strings"

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

		// TODO: optimize this chunk. do not copy full line.
		line.Value = make([]rune, len(prevLine.Value))
		copy(line.Value, prevLine.Value)
		CmdCursorFirstNonBlank(ctx)
		line.Value = line.Value[:ctx.Buf.Cursor.Char]

		if prevLine.Value[len(prevLine.Value)-1] == '{' {
			line.Value = append(line.Value, '\t')
			CmdCursorRight(ctx)
		}
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
