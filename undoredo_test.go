package wig

import (
	"bytes"
	"slices"
	"testing"

	"github.com/firstrow/wig/testutils"
	"github.com/stretchr/testify/assert"
)

func TestEdits(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	buf := e.OpenFile(testutils.Filepath("buffer_test.txt"))
	buf.Cursor.Char = 0
	buf.Cursor.Line = 0
	buf.Selection = nil

	dupLines := List[Line]{}

	// dup lines
	cl := buf.Lines.First()
	for cl != nil {
		tmpData := slices.Clone(cl.Value)
		dupLines.PushBack(Line(tmpData))
		cl = cl.Next()
	}

	assert.Equal(t, linesToString(dupLines), buf.String())
}

func linesToString(l List[Line]) string {
	buf := bytes.NewBuffer(nil)

	line := l.First()
	for line != nil {
		buf.WriteString(string(line.Value))
		line = line.Next()
	}

	return buf.String()
}

