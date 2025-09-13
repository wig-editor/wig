package wig

func TextObjectWord(ctx Context, bigword bool) (start, end int) {
	cur := ContextCursorGet(ctx)
	start = cur.Char
	end = start

	line := CursorLine(ctx.Buf, cur)
	cls := CursorChClass(ctx.Buf, cur)

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
// TODO: rewrite
func TextObjectBlock(buf *Buffer, ch rune, include bool) (found bool, sel *Selection, cur Cursor) {
	// defer func(c Cursor) {
	// buf.Cursor = c
	// }(buf.Cursor)

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
		if CursorChar(buf, nil) == openCh {
			openChFound = true
			break
		}
		if !CursorDec(buf, nil) {
			break
		}
	}
	if !openChFound {
		return
	}

	// TODO: fix
	bufCursor := Cursor{}
	start := bufCursor

	// move cursor "left" till we find first "open" bracket
	for {
		if CursorChar(buf, nil) == closeCh {
			end := bufCursor
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
			}, bufCursor
		}
		if !CursorInc(buf, nil) {
			break
		}
	}

	return false, nil, bufCursor
}

