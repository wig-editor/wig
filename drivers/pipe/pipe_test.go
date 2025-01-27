package pipe

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/firstrow/mcwig"
	"github.com/firstrow/mcwig/testutils"
)

func TestPipe(t *testing.T) {
	e := mcwig.NewEditor(
		testutils.Viewport,
		nil,
	)

	buf := e.BufferFindByFilePath("test-1", true)
	lines := `#
#  cmd: echo "%s"
#
`
	for _, l := range strings.Split(lines, "\n") {
		buf.Append(l)
	}

	buf.Cursor.Line = 2

	p := New(e)
	h := p.parseHeader(buf)
	assert.Equal(t, "echo", p.getCommand(h.cmd))

	outBuf := e.BufferFindByFilePath("output", true)
	p.send(h, outBuf, `ping pong`)
	p.cmd.Wait()

	assert.Equal(t, "ping pong", outBuf.String())

	args := p.buildArgs(h.cmd, "ping pong")
	assert.Equal(t, 1, len(args))
	assert.Equal(t, `ping pong`, args[0])
}

func Test_LongRunningProcess(t *testing.T) {
	e := mcwig.NewEditor(
		testutils.Viewport,
		nil,
	)

	buf := e.BufferFindByFilePath("test-1", true)
	lines := `#
# cmd: python -i
# interactive: 1
# append: 1
#
`
	for _, l := range strings.Split(lines, "\n") {
		buf.Append(l)
	}

	buf.Cursor.Line = 4
	p := New(e)
	h := p.parseHeader(buf)

	assert.Equal(t, "python", p.getCommand(h.cmd))
	outBuf := e.BufferFindByFilePath("output", true)
	p.send(h, outBuf, `help`)
	// TODO: figure out how to Wait properly
	time.Sleep(100 * time.Millisecond)

	assert.Contains(t, outBuf.String(), "Type help() for interactive help, or help(object) for help about object.")
}

func Test_ParseHeader(t *testing.T) {
	e := mcwig.NewEditor(
		testutils.Viewport,
		nil,
	)

	lines := `#
#  cmd   : bin --arg=1 a:b
#interactive: t
#    append: false
#

hello world

#  cmd: uname

-a
`

	buf := e.BufferFindByFilePath("test-1", true)
	for _, l := range strings.Split(lines, "\n") {
		buf.Append(l)
	}

	buf.Cursor.Line = 5
	p := New(e)

	result := p.parseHeader(buf)
	expected := header{
		cmd:         "bin --arg=1 a:b",
		interactive: true,
		append:      false,
	}
	assert.Equal(t, expected, result)

	// test "down-up" parsing
	buf.Cursor.Line = 11

	result = p.parseHeader(buf)
	expected = header{
		cmd:         "uname",
		interactive: false,
		append:      false,
	}
	assert.Equal(t, expected, result)
	assert.Equal(t, expected, result)
}
