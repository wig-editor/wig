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

// Returns selection inside '(', '{', '[' as "Selection" range. This implementation is simple
// and does not check if open/close symbols are "balanced".
func TextObjectBlock(buf *Buffer, ch rune, include bool) (found bool, sel *Selection, cur Cursor) {
	defer func(c Cursor) {
		buf.Cursor = c
	}(buf.Cursor)

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

	// move cursor "left" till we find first "open" bracket
	for {
		if CursorChar(buf) == closeCh {
			end := buf.Cursor
			if include == false {
				// no selection. empty ().
				if end.Char == start.Char+1 {
					return true, nil, end
				}

				start.Char += 1
				end.Char -= 1
			}

			return true, &Selection{
				Start: start,
				End:   end,
			}, buf.Cursor
		}
		if !CursorInc(buf) {
			break
		}
	}

	return false, nil, buf.Cursor
}
