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
	cursor *Element[HighlighterNode]
}

func HighlighterGet(b *Buffer) Highlighter {
	// TODO: implement
	return nil
}

func (c *HighlighterCursor) Seek(line, ch uint32) (node *Element[HighlighterNode], found bool) {
	inRange := func(node *Element[HighlighterNode], line, ch uint32) bool {
		if node == nil {
			return false
		}
		if line >= node.Value.StartLine && line <= node.Value.EndLine {
			if ch >= node.Value.StartChar && ch < node.Value.EndChar {
				return true
			}
		}
		return false
	}

	if inRange(c.cursor, line, ch) {
		return c.cursor, true
	}

	nextNode := c.cursor.Next()
	for nextNode != nil {
		if nextNode.Value.StartLine > line {
			break
		}

		if inRange(nextNode, line, ch) {
			c.cursor = nextNode
			return c.cursor, true
		}

		nextNode = nextNode.Next()
	}

	return nil, false
}

