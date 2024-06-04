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

	buf := e.BufferGetByName("test-1")
	lines := `#
#  cmd: echo "%s"
#
`
	for _, l := range strings.Split(lines, "\n") {
		buf.AppendStringLine(l)
	}

	p := New(e)
	h := p.parseHeader(buf)
	assert.Equal(t, "echo", p.getCommand(h.cmd))

	outBuf := e.BufferGetByName("output")
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

	buf := e.BufferGetByName("test-1")
	lines := `#
# cmd: python -i
# interactive: 1
# append: 1
#
`
	for _, l := range strings.Split(lines, "\n") {
		buf.AppendStringLine(l)
	}

	p := New(e)
	h := p.parseHeader(buf)

	assert.Equal(t, "python", p.getCommand(h.cmd))
	outBuf := e.BufferGetByName("output")
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
`

	buf := e.BufferGetByName("test-1")
	for _, l := range strings.Split(lines, "\n") {
		buf.AppendStringLine(l)
	}

	p := New(e)
	result := p.parseHeader(buf)

	expected := header{
		cmd:         "bin --arg=1 a:b",
		interactive: true,
		append:      false,
	}

	assert.Equal(t, expected, result)
}
