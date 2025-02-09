package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/require"
)

func TestBuildTextChangeEvent(t *testing.T) {
	source := `package mcwig

import "fmt"

func add(a int, b int) {
	fmt.Printf("%d", a+b)
}`

	e := NewEditor(
		testutils.Viewport,
		nil,
	)
	buf := e.BufferFindByFilePath("testfile", true)
	buf.ResetLines()
	buf.Append(source)
	require.Equal(t, source+"\n", buf.String())

	events := e.Events.Subscribe()

	line := CursorLineByNum(buf, 4)
	TextInsert(buf, line, 22, " int")
	require.Equal(t, "func add(a int, b int) int {\n", line.Value.String())

	msg := <-events
	event := msg.(EventTextChange)
	require.Equal(t, EventTextChange{
		Buf:   buf,
		Start: Position{Line: 4, Char: 22},
		End:   Position{Line: 4, Char: 22},
		Text:  " int",
	}, event)
}
