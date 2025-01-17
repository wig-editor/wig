package mcwig

import (
	"bytes"
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEdits(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")
	buf := e.ActiveBuffer()
	buf.Cursor.Char = 0
	buf.Cursor.Line = 0
	buf.Selection = nil

	dupLines := List[Line]{}

	// dup lines
	cl := buf.Lines.First()
	for cl != nil {
		tmpData := make([]rune, len(cl.Value))
		copy(tmpData, cl.Value)
		dupLines.PushBack(Line(tmpData))
		cl = cl.Next()
	}

	assert.Equal(t, linesToString(dupLines), buf.String())

	panic(1)
}

func linesToString(l List[Line]) string {
	buf := bytes.NewBuffer(nil)

	line := l.First()
	sep := "\n"
	for line != nil {
		if line.Next() == nil {
			sep = ""
		}
		buf.WriteString(string(line.Value) + sep)
		line = line.Next()
	}

	return buf.String()
}
