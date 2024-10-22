package mcwig

// Aligns with previous non-empty line. Also checks for [{:
func CmdIndent(e *Editor) {
	Do(e, func(buf *Buffer, line *Element[Line]) {
		prevLine := line.Prev()
		for prevLine != nil {
			if prevLine.Value.IsEmpty() {
				prevLine = prevLine.Prev()
				continue
			}

			line.Value = make([]rune, len(prevLine.Value))
			copy(line.Value, prevLine.Value)
			CmdCursorFirstNonBlank(e)
			line.Value = line.Value[:buf.Cursor.Char]

			// TODO: check range if chars
			if prevLine.Value[len(prevLine.Value)-1] == '{' {
				// increase indentation
				line.Value = append(line.Value, '	')
			}
			break
		}
	})
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
