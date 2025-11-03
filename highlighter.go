package wig

type Highlighter interface {
	// Full document build/re-build/init
	Build()
	TextChanged(EventTextChange)

	// Query syntax highlights for given range
	ForRange(startLine, endLine uint32) *HighlighterCursor
}

type HighlighterNode struct {
	NodeName  string
	StartLine uint32
	StartChar uint32
	EndLine   uint32
	EndChar   uint32
}

type HighlighterCursor struct {
	Cursor *Element[HighlighterNode]
}

func (c *HighlighterCursor) Seek(line, ch uint32) (node *Element[HighlighterNode], found bool) {
	if c.Cursor == nil {
		return
	}

	inRange := func(node *Element[HighlighterNode], line, ch uint32) bool {
		if node == nil {
			return false
		}
		if line >= node.Value.StartLine && line <= node.Value.EndLine {
			if node.Value.EndLine > line {
				return true
			}
			if ch >= node.Value.StartChar && ch < node.Value.EndChar {
				return true
			}
		}
		return false
	}

	if inRange(c.Cursor, line, ch) {
		return c.Cursor, true
	}

	nextNode := c.Cursor.Next()

	for nextNode != nil {
		if nextNode.Value.StartLine > line {
			break
		}

		if inRange(nextNode, line, ch) {
			c.Cursor = nextNode
			return c.Cursor, true
		}

		nextNode = nextNode.Next()
	}

	return nil, false
}

