package mcwig

func TextObjectWord(buf *Buffer, bigword bool) (start, end int) {
	start = buf.Cursor.Char
	end = start

	line := CursorLine(buf)
	cls := CursorChClass(buf)

	if bigword {
		for start > 0 {
			if line.Value.IsEmpty() {
				break
			}
			if getChClass(line.Value[start-1]) == cls {
				start--
			} else {
				break
			}
		}
	}

	end = start

	for i, r := range line.Value {
		if i < start {
			continue
		}
		if getChClass(r) == cls {
			end = i
			continue
		}

		break
	}

	return start, end
}

// Returns text inside '(', '{', '[' as Selection range. This implementation is simple
// and does not check if open/close symbols are "balanced".
func TextObjectBlock(buf *Buffer, ch rune, include bool) (found bool, sel Selection) {
	cursor := buf.Cursor
	// restore cursor position on exit
	defer func(c Cursor) {
		buf.Cursor = c
	}(cursor)

	openClose := map[rune]rune{
		'(': ')',
		'{': '}',
		'[': ']',
	}

	if _, ok := openClose[ch]; !ok {
		for k, v := range openClose {
			if v == ch {
				ch = k
				break
			}
		}
	}
	openCh := ch
	closeCh := openClose[ch]

	// move cursor back until 'openCh' is found
	openChFound := false
	for {
		if CursorChar(buf) == openCh {
			openChFound = true
			break
		}
		if !CursorDec(buf) {
			break
		}
	}
	if !openChFound {
		return
	}

	start := buf.Cursor

	for {
		if CursorChar(buf) == closeCh {
			return true, Selection{
				Start: start,
				End:   buf.Cursor,
			}
		}
		if !CursorInc(buf) {
			break
		}
	}

	return
}
