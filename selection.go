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
