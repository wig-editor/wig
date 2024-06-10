package mcwig

type Selection struct {
	Start Cursor
	End   Cursor
}

func SelectionCursorInRange(sel *Selection, c Cursor) bool {
	s := *sel

	if s.Start.Line > s.End.Line {
		s.Start, s.End = s.End, s.Start
	}

	if s.Start.Line == s.End.Line {
		if s.Start.Char > s.End.Char {
			s.Start, s.End = s.End, s.Start
		}
	}

	if c.Line < s.Start.Line || c.Line > s.End.Line {
		return false
	}

	if c.Line == s.Start.Line {
		if c.Char < s.Start.Char {
			return false
		}
	}

	if c.Line == s.End.Line {
		if c.Char > s.End.Char {
			return false
		}
	}

	return true
}

// FIXME: make it right. dammit.
func SelectionToString(buf *Buffer) string {
	if buf.Selection == nil {
		return ""
	}

	s := *buf.Selection

	if s.Start.Line > s.End.Line {
		s.Start, s.End = s.End, s.Start
	}

	if s.Start.Line == s.End.Line {
		if s.Start.Char > s.End.Char {
			s.Start, s.End = s.End, s.Start
		}
	}

	lineStart := lineByNum(buf, s.Start.Line)
	lineEnd := lineByNum(buf, s.End.Line)

	if lineStart == nil {
		return ""
	}

	if s.Start.Line == s.End.Line {
		chLen := s.End.Char + 1
		if chLen > len(lineStart.Value) {
			chLen = len(lineStart.Value) - 1
		}
		if chLen <= 0 {
			return ""
		}
		return string(lineStart.Value[s.Start.Char:chLen])
	}

	result := string(lineStart.Value[s.Start.Char:])
	currentLine := lineStart.Next()
	for currentLine != nil {
		if currentLine != lineEnd {
			result += "\n" + string(currentLine.Value)
		} else {
			chLen := s.End.Char + 1
			if chLen > len(currentLine.Value) {
				chLen = len(currentLine.Value) - 1
			}
			if chLen <= 0 {
				return result
			}
			result += "\n" + string(currentLine.Value[:chLen])
			break
		}
		currentLine = currentLine.Next()
	}

	return result
}
